#!/usr/bin/env bash
# validate-e2e-dev.sh
#
# Variante del validate-e2e-local apuntando al ALB público Dev a través
# del api-gateway. Todas las llamadas pasan por:
#   https://api-dev.farmanexo.com.pe (cuenta AWS 041277794550, us-east-1)
#
# Mismos 11 steps que la versión local. Para usar:
#   bash validate-e2e-dev.sh
set -euo pipefail

GW=https://api-dev.farmanexo.com.pe

# En Dev solo hay un origen público (api-gateway). Mantengo nombres para
# que el resto del script sea idéntico al local.
AUTH=$GW
USERSVC=$GW
CATALOG=$GW
PHARMACY=$GW
ORDER=$GW

TS=$(date +%s)
EMAIL="e2e_dev_${TS}@farmanexo.com"
PASS="TestPassword123!"

OUT_DIR=$(mktemp -d)
trap 'rm -rf "$OUT_DIR"' EXIT
LAST="$OUT_DIR/last.json"

step=0
say()  { echo -e "\n\033[1;36m── STEP $step ── $*\033[0m"; }
ok()   { echo -e "\033[1;32m✅ $*\033[0m"; }
fail() { echo -e "\033[1;31m❌ $*\033[0m"; exit 1; }

pyget() {
  cat "$LAST" | python -c "import json,sys; d=json.load(sys.stdin); v=$1; print('null' if v is None else v)"
}

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

step=1; say "Register usuario LPDP-compliant"
req POST "$AUTH/api/v1/auth/register" "{
  \"email\":\"$EMAIL\",
  \"password\":\"$PASS\",
  \"full_name\":\"Usuario E2E Dev\",
  \"phone\":\"999888777\",
  \"accepted_terms\":true,
  \"accepted_privacy\":true,
  \"marketing_opt_in\":false
}"
USER_ID=$(pyget "d['datos']['user_id']")
ok "user_id=$USER_ID"

step=2; say "Login → access_token"
req POST "$AUTH/api/v1/auth/login" "{\"email\":\"$EMAIL\",\"password\":\"$PASS\"}"
ACCESS=$(pyget "d['datos']['access_token']")
[[ -n "$ACCESS" && "$ACCESS" != "null" ]] || fail "access_token vacío"
ok "JWT obtenido (len=${#ACCESS})"

step=3; say "Esperar a que user-service consuma USER_REGISTERED"
for i in {1..30}; do
  http=$(curl -s -o "$LAST" -w "%{http_code}" \
    -H "Authorization: Bearer $ACCESS" "$USERSVC/api/v1/users/me" || echo 000)
  [[ "$http" = "200" ]] && { ok "profile listo tras ${i}s"; break; }
  sleep 1
  [[ $i = 30 ]] && fail "profile no apareció en 30s (SQS consumer down?)"
done

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
ok "address_id=$ADDRESS_ID"

step=5; say "Tomar 1 producto real DIGEMID"
req GET "$CATALOG/api/v1/products?limit=1"
PRODUCT_ID=$(pyget "d['datos']['products'][0]['id']")
PRODUCT_NAME=$(pyget "d['datos']['products'][0]['name']")
ok "product=$PRODUCT_NAME ($PRODUCT_ID)"

step=6; say "Encontrar farmacia con stock > 0"
req GET "$PHARMACY/api/v1/pharmacies/inventory/product/$PRODUCT_ID"
PHARMACY_ID=$(pyget "next(i for i in d['datos']['items'] if i['stock']>0)['pharmacy_id']")
PHARMACY_NAME=$(pyget "next(i for i in d['datos']['items'] if i['stock']>0)['pharmacy_name']")
UNIT_PRICE=$(pyget "next(i for i in d['datos']['items'] if i['stock']>0)['price']")
ok "pharmacy=$PHARMACY_NAME ($PHARMACY_ID)  precio=S/$UNIT_PRICE"

step=7; say "POST /cart/items (K16 path en Dev)"
req POST "$ORDER/api/v1/cart/items" "{
  \"product_id\":\"$PRODUCT_ID\",
  \"pharmacy_id\":\"$PHARMACY_ID\",
  \"quantity\":1
}" "$ACCESS"
ITEMS=$(pyget "len(d['datos']['items'])")
CART_ITEM_ID=$(pyget "d['datos']['items'][0]['id']")
ok "items=$ITEMS  cart_item_id=$CART_ITEM_ID"

step=8; say "GET /cart → verificar total"
req GET "$ORDER/api/v1/cart" "" "$ACCESS"
TOTAL=$(pyget "d['datos']['total_amount']")
ok "cart total=S/$TOTAL"

step=9; say "POST /orders/checkout (K18 path en Dev)"
req POST "$ORDER/api/v1/orders/checkout" "{
  \"delivery_method\":\"delivery\",
  \"delivery_address_id\":\"$ADDRESS_ID\",
  \"payment_method\":\"card\",
  \"payment_details\":{\"card_token\":\"tok_test_e2e\"},
  \"notes\":\"E2E Dev validation\"
}" "$ACCESS"
ORDER_ID=$(pyget "d['datos']['orders'][0]['order_id']")
ORDER_NUMBER=$(pyget "d['datos']['orders'][0]['order_number']")
ORDER_STATUS=$(pyget "d['datos']['orders'][0]['status']")
ok "order=$ORDER_NUMBER status=$ORDER_STATUS id=$ORDER_ID"

step=10; say "GET /orders/{id} → snapshot + status history"
req GET "$ORDER/api/v1/orders/$ORDER_ID" "" "$ACCESS"
ITEMS_COUNT=$(pyget "len(d['datos'].get('items') or [])")
HISTORY_COUNT=$(pyget "len(d['datos'].get('status_history') or [])")
ok "items=$ITEMS_COUNT history_entries=$HISTORY_COUNT"

step=11; say "POST /orders/{id}/cancel → ORDER_CANCELLED"
req POST "$ORDER/api/v1/orders/$ORDER_ID/cancel" \
  '{"reason":"E2E Dev test cleanup"}' "$ACCESS"
ok "orden cancelada"

echo
echo -e "\033[1;32m═══ DEV HAPPY PATH OK ═══\033[0m"
echo "  USER_ID:       $USER_ID"
echo "  EMAIL:         $EMAIL"
echo "  ADDRESS_ID:    $ADDRESS_ID"
echo "  PRODUCT:       $PRODUCT_NAME"
echo "  PHARMACY:      $PHARMACY_NAME"
echo "  ORDER_NUMBER:  $ORDER_NUMBER"
echo "  ORDER_ID:      $ORDER_ID"
