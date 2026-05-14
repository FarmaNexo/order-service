#!/usr/bin/env bash
# validate-e2e-local.sh
#
# Happy path order flow contra localhost — valida K15 (CATALOG_SERVICE_URL,
# PHARMACY_SERVICE_URL, USER_SERVICE_URL) en runtime con datos DIGEMID reales.
#
# Requiere: 4001 auth, 4002 user, 4003 catalog, 4004 pharmacy, 4011 order up.
# Usa python (no jq) para parsing — más portable en Git Bash Windows.
set -euo pipefail

AUTH=http://localhost:4001
USERSVC=http://localhost:4002
CATALOG=http://localhost:4003
PHARMACY=http://localhost:4004
ORDER=http://localhost:4011

TS=$(date +%s)
EMAIL="e2e_order_${TS}@farmanexo.com"
PASS="TestPassword123!"

OUT_DIR=$(mktemp -d)
trap 'rm -rf "$OUT_DIR"' EXIT
LAST="$OUT_DIR/last.json"

step=0
say()  { echo -e "\n\033[1;36m── STEP $step ── $*\033[0m"; }
ok()   { echo -e "\033[1;32m✅ $*\033[0m"; }
fail() { echo -e "\033[1;31m❌ $*\033[0m"; exit 1; }

# pyget EXPR — eval Python expression with `d` = parsed last.json (via stdin).
# Single-line python -c (pyenv-win shim corrupts multiline strings).
pyget() {
  cat "$LAST" | python -c "import json,sys; d=json.load(sys.stdin); v=$1; print('null' if v is None else v)"
}

# req METHOD URL [BODY] [BEARER] — fail-fast on non-2xx, dumps body
req() {
  local method=$1 url=$2 body=${3:-} bearer=${4:-}
  local -a curl_args=(-sS -X "$method" -o "$LAST"
                      -w "%{http_code} %{time_total}"
                      -H "Content-Type: application/json"
                      -H "Accept: application/json")
  [[ -n "$bearer" ]] && curl_args+=(-H "Authorization: Bearer $bearer")
  [[ -n "$body" ]]   && curl_args+=(--data "$body")
  local resp status time
  resp=$(curl "${curl_args[@]}" "$url" 2>/dev/null || echo "000 0")
  status=${resp%% *}
  time=${resp##* }
  if [[ ! "$status" =~ ^2 ]]; then
    echo "  → $method $url"
    echo "  → status=$status latency=${time}s"
    echo "  → body:"
    python -m json.tool "$LAST" 2>/dev/null || cat "$LAST"
    fail "HTTP $status"
  fi
  printf "   %s %s → \033[32m%s\033[0m (%ss)\n" "$method" "$url" "$status" "$time"
}

# ─── STEP 1 ────────────────────────────────────────────────────────────────
step=1; say "Register usuario LPDP-compliant"
req POST "$AUTH/api/v1/auth/register" "{
  \"email\":\"$EMAIL\",
  \"password\":\"$PASS\",
  \"full_name\":\"Usuario E2E\",
  \"phone\":\"999888777\",
  \"accepted_terms\":true,
  \"accepted_privacy\":true,
  \"marketing_opt_in\":false
}"
USER_ID=$(pyget "d['datos']['user_id']")
ok "user_id=$USER_ID"

# ─── STEP 2 ────────────────────────────────────────────────────────────────
step=2; say "Login → access_token"
req POST "$AUTH/api/v1/auth/login" "{\"email\":\"$EMAIL\",\"password\":\"$PASS\"}"
ACCESS=$(pyget "d['datos']['access_token']")
[[ -n "$ACCESS" && "$ACCESS" != "null" ]] || fail "access_token vacío"
ok "JWT obtenido (len=${#ACCESS})"

# ─── STEP 3 ────────────────────────────────────────────────────────────────
step=3; say "Esperar a que user-service consuma USER_REGISTERED (poll profile)"
for i in {1..15}; do
  http=$(curl -s -o "$LAST" -w "%{http_code}" \
    -H "Authorization: Bearer $ACCESS" "$USERSVC/api/v1/users/me" || echo 000)
  [[ "$http" = "200" ]] && { ok "profile listo tras ${i}s"; break; }
  sleep 1
  [[ $i = 15 ]] && fail "profile no apareció en 15s (SQS consumer down?)"
done

# ─── STEP 4 ────────────────────────────────────────────────────────────────
step=4; say "Crear address Miraflores"
req POST "$USERSVC/api/v1/users/me/addresses" '{
  "label":"Casa",
  "street":"Av. Larco 1100",
  "city":"MIRAFLORES",
  "state":"Lima",
  "postal_code":"15074",
  "country":"PE",
  "is_default":true,
  "latitude":-12.1267,
  "longitude":-77.0290
}' "$ACCESS"
ADDRESS_ID=$(pyget "d['datos']['id']")
[[ -n "$ADDRESS_ID" && "$ADDRESS_ID" != "null" ]] || fail "address_id vacío"
ok "address_id=$ADDRESS_ID"

# ─── STEP 5 ────────────────────────────────────────────────────────────────
step=5; say "Tomar 1 producto real DIGEMID"
req GET "$CATALOG/api/v1/products?limit=1"
PRODUCT_ID=$(pyget "d['datos']['products'][0]['id']")
PRODUCT_NAME=$(pyget "d['datos']['products'][0]['name']")
ok "product=$PRODUCT_NAME ($PRODUCT_ID)"

# ─── STEP 6 ────────────────────────────────────────────────────────────────
step=6; say "Encontrar farmacia con stock > 0"
req GET "$PHARMACY/api/v1/pharmacies/inventory/product/$PRODUCT_ID"
PHARMACY_ID=$(pyget "next(i for i in d['datos']['items'] if i['stock']>0)['pharmacy_id']")
PHARMACY_NAME=$(pyget "next(i for i in d['datos']['items'] if i['stock']>0)['pharmacy_name']")
UNIT_PRICE=$(pyget "next(i for i in d['datos']['items'] if i['stock']>0)['price']")
[[ -n "$PHARMACY_ID" && "$PHARMACY_ID" != "null" ]] || fail "ninguna farmacia con stock"
ok "pharmacy=$PHARMACY_NAME ($PHARMACY_ID)  precio=S/$UNIT_PRICE"

# ─── STEP 7 ────────────────────────────────────────────────────────────────
step=7; say "POST /cart/items (order → catalog + order → pharmacy)"
req POST "$ORDER/api/v1/cart/items" "{
  \"product_id\":\"$PRODUCT_ID\",
  \"pharmacy_id\":\"$PHARMACY_ID\",
  \"quantity\":1
}" "$ACCESS"
ITEMS=$(pyget "len(d['datos']['items'])")
CART_ITEM_ID=$(pyget "d['datos']['items'][0]['id']")
ok "items=$ITEMS  cart_item_id=$CART_ITEM_ID"

# ─── STEP 8 ────────────────────────────────────────────────────────────────
step=8; say "GET /cart → verificar snapshot poblado"
req GET "$ORDER/api/v1/cart" "" "$ACCESS"
TOTAL=$(pyget "d['datos']['total_amount']")
ok "cart total=S/$TOTAL"

# ─── STEP 9 ────────────────────────────────────────────────────────────────
step=9; say "POST /orders/checkout (order → user para address)"
req POST "$ORDER/api/v1/orders/checkout" "{
  \"delivery_method\":\"delivery\",
  \"delivery_address_id\":\"$ADDRESS_ID\",
  \"payment_method\":\"card\",
  \"payment_details\":{\"card_token\":\"tok_test_e2e\"},
  \"notes\":\"E2E validation\"
}" "$ACCESS"
ORDER_ID=$(pyget "d['datos']['orders'][0]['order_id']")
ORDER_NUMBER=$(pyget "d['datos']['orders'][0]['order_number']")
ORDER_STATUS=$(pyget "d['datos']['orders'][0]['status']")
[[ -n "$ORDER_ID" && "$ORDER_ID" != "null" ]] || fail "checkout no devolvió order_id"
ok "order=$ORDER_NUMBER status=$ORDER_STATUS id=$ORDER_ID"

# ─── STEP 10 ───────────────────────────────────────────────────────────────
step=10; say "GET /orders/{id} → verificar snapshot + status history"
req GET "$ORDER/api/v1/orders/$ORDER_ID" "" "$ACCESS"
ITEMS_COUNT=$(pyget "len(d['datos'].get('items') or [])")
HISTORY_COUNT=$(pyget "len(d['datos'].get('status_history') or [])")
ok "items=$ITEMS_COUNT history_entries=$HISTORY_COUNT"

# ─── STEP 11 (bonus) ───────────────────────────────────────────────────────
step=11; say "POST /orders/{id}/cancel → debe publicar ORDER_CANCELLED"
req POST "$ORDER/api/v1/orders/$ORDER_ID/cancel" \
  '{"reason":"E2E test cleanup"}' "$ACCESS"
ok "orden cancelada"

# ─── DONE ──────────────────────────────────────────────────────────────────
echo
echo -e "\033[1;32m═══ HAPPY PATH OK ═══\033[0m"
echo "  USER_ID:       $USER_ID"
echo "  EMAIL:         $EMAIL"
echo "  ADDRESS_ID:    $ADDRESS_ID"
echo "  PRODUCT:       $PRODUCT_NAME"
echo "  PHARMACY:      $PHARMACY_NAME"
echo "  ORDER_NUMBER:  $ORDER_NUMBER"
echo "  ORDER_ID:      $ORDER_ID"
