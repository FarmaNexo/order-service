# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Project Overview

FarmaNexo Order Service — a Go microservice handling **shopping cart, checkout, orders, and payments** for the FarmaNexo marketplace. Clean Architecture + CQRS + MediatR pattern.

**Scope**: Multi-pharmacy shopping cart, checkout with order splitting per pharmacy, payment processing (mock), order lifecycle management, pharmacy order management, admin statistics, SQS consumer for INVENTORY_UPDATED/PRODUCT_UPDATED events, SQS publisher for ORDER events, Redis caching.

## Common Commands

```bash
make build             # Compile binary to bin/order-service
make dev               # Run locally (ENV=local) with swagger generation
make run               # Run with swagger (uses ENV variable, default: local)
make test              # Run all tests with race detection
make lint              # Run golangci-lint
make swagger           # Generate Swagger docs
make migrate-up        # Apply pending DB migrations
make migrate-down      # Revert last migration
make migrate-create NAME=xxx  # Create new migration pair
```

## Architecture

**Clean Architecture + CQRS** with a custom Mediator pattern.

### Layer Structure (`internal/`)

- **domain/** — Entities (`CartItem`, `Order`, `OrderItem`, `OrderStatusHistory`), repository interfaces (4), service interfaces (CacheService, EventPublisher, JWTService, PharmacyClient, CatalogClient, UserClient, PaymentService). No external dependencies.
- **application/** — Commands (7), Queries (6), handlers (13 total + 1 helper), validators (4), preprocessor (1), postprocessor (1).
- **infrastructure/** — PostgreSQL repos (GORM), JWT validation (`golang-jwt/v5`), SQS publisher + consumer, Redis cache, HTTP clients (PharmacyClient, CatalogClient, UserClient), Mock Payment Service.
- **presentation/** — Chi HTTP router, controllers, DTOs (request/response), middlewares (auth JWT + admin role + pharmacy_owner, correlation ID).
- **shared/** — Cross-cutting: `ApiResponse[T]`, domain errors, constants.

### Request Flow

```
HTTP → Chi Router → [AuthMiddleware] → Controller → Mediator.Send(Command/Query) → Validation → PreProcessor → Handler → PostProcessor → ApiResponse
```

### Key Design Decision: Multi-Pharmacy Cart

The cart allows products from MULTIPLE pharmacies in a single session. At checkout:
- Items are grouped by pharmacy_id
- ONE Order is created per pharmacy
- Each pharmacy receives only their order
- Order numbers are generated sequentially (ORD-YYYY-NNNNNN)
- Payment is processed once for the total amount
- On success: all orders confirmed, cart cleared, events published
- On failure: orders remain as pending_payment, cart NOT cleared

## API Endpoints

### Cart (Authenticated - requires Bearer JWT)
| Endpoint | Method | Handler | Description |
|---|---|---|---|
| `/api/v1/cart` | GET | GetCartHandler | Get current cart with pharmacy grouping |
| `/api/v1/cart/items` | POST | AddCartItemHandler | Add product to cart (validates inventory) |
| `/api/v1/cart/items/{item_id}` | PUT | UpdateCartItemHandler | Update item quantity |
| `/api/v1/cart/items/{item_id}` | DELETE | RemoveCartItemHandler | Remove item from cart |
| `/api/v1/cart` | DELETE | ClearCartHandler | Empty entire cart |

### Orders (Authenticated)
| Endpoint | Method | Handler | Description |
|---|---|---|---|
| `/api/v1/orders/checkout` | POST | CheckoutHandler | Create orders from cart + process payment |
| `/api/v1/orders` | GET | ListMyOrdersHandler | List user's orders (paginated) |
| `/api/v1/orders/{order_id}` | GET | GetOrderDetailHandler | Full order detail with items + status history |
| `/api/v1/orders/{order_id}/cancel` | POST | CancelOrderHandler | Cancel order (only pending_payment/confirmed) |

### Pharmacy (requires pharmacy_owner role)
| Endpoint | Method | Handler | Description |
|---|---|---|---|
| `/api/v1/pharmacy/orders` | GET | ListPharmacyOrdersHandler | List pharmacy's orders |
| `/api/v1/pharmacy/orders/{order_id}/status` | PUT | UpdateOrderStatusHandler | Update order status |

### Admin (requires admin role)
| Endpoint | Method | Handler | Description |
|---|---|---|---|
| `/api/v1/admin/orders` | GET | ListAllOrdersHandler | All orders with filters |
| `/api/v1/admin/orders/stats` | GET | GetOrderStatsHandler | Order statistics |

### Health (No auth required)
| Endpoint | Method | Handler | Description |
|---|---|---|---|
| `/health` | GET | HealthCheck | Service health (ApiResponse[T]) |
| `/` | GET | HealthCheck | Service health (alias) |
| `/api/v2/status` | GET | HealthCheck | API v2 status |

## Database

- **Database**: `order_db`
- **Schema**: `orders`
- **Tables**: `cart_items`, `orders`, `order_items`, `order_status_history`
- **Sequence**: `order_number_seq` for unique order numbers
- **Migrations**: `migrations/` using golang-migrate

## Order Status Flow

```
pending_payment → confirmed → preparing → ready → in_delivery → completed
       ↓              ↓
    cancelled      cancelled
```

## Event System

### Consumed (from SQS)
- `INVENTORY_UPDATED` ← Queue: farmanexo-pharmacy-events
- `PRODUCT_UPDATED` ← Queue: farmanexo-catalog-events

### Published (to SQS)
- `ORDER_CREATED` → Queue: farmanexo-order-events
- `ORDER_CONFIRMED` → Queue: farmanexo-order-events
- `ORDER_STATUS_UPDATED` → Queue: farmanexo-order-events
- `ORDER_COMPLETED` → Queue: farmanexo-order-events
- `ORDER_CANCELLED` → Queue: farmanexo-order-events
- `PAYMENT_COMPLETED` → Queue: farmanexo-order-events
- `PAYMENT_FAILED` → Queue: farmanexo-order-events

## Inter-Service Communication

### HTTP Clients (3s timeout)
- **Pharmacy Service** (port 4004): Verify inventory, get pharmacy info
- **Catalog Service** (port 4003): Get product info
- **User Service** (port 4002): Validate delivery addresses

## User Stories Implemented

- **HU-043**: Carrito de compras multi-farmacia
- **HU-044**: Proceso de pago (mock payment service)
- **HU-045**: Notificación a farmacias (via SQS events)
- **HU-046**: Gestión de órdenes (CRUD + status management)

## Known Issues

### K16 — order→pharmacy GetInventoryItem hits non-existent endpoint (detectado 2026-05-13)

`internal/infrastructure/clients/pharmacy_client_impl.go:28` llama:
```
GET {PHARMACY}/api/v1/pharmacies/{pharmacyID}/inventory/{productID}
```
**Pero pharmacy-service NO expone esa ruta como GET.** Solo existen:
- `GET /api/v1/pharmacies/{id}/inventory` (todo el inventario de una farmacia)
- `GET /api/v1/pharmacies/inventory/product/{productId}` (cross-pharmacy comparator)
- `PUT/DELETE /api/v1/pharmacies/{id}/inventory/{productId}` (admin owner)

**Síntoma:** pharmacy responde `405 Method Not Allowed` (Allow: PUT, DELETE). order trata cualquier non-200 como error → response `400 BUS_001 "Producto no disponible en esta farmacia"`. Bloquea `POST /cart/items` para CUALQUIER combinación de producto+farmacia, incluso con stock disponible.

**Reproducción** (con stack local up): `bash services/order-service/scripts/validate-e2e-local.sh` falla en STEP 7.

**Scope:** afecta cart-add + posiblemente checkout (mismo client). Lado código (NO IaC).

**Opciones de fix** (decidir antes de implementar):
- (A) Agregar `GET /api/v1/pharmacies/{id}/inventory/{productId}` en pharmacy-service (handler nuevo + route). Más limpio, sigue REST.
- (B) Cambiar `pharmacy_client_impl.go` para llamar `GET /api/v1/pharmacies/inventory/product/{productId}` y filtrar por `pharmacy_id` en el response.items[]. Una llamada de red potencialmente más grande pero reutiliza endpoint existente.

**Recomendado: (A)** — el endpoint cross-pharmacy es para comparador público, no para validar 1 item. Separar concerns. Migration de código sola (no schema).

**Estado:** ✅ resuelto 2026-05-13 (opción A). Pharmacy expone `GET /api/v1/pharmacies/{id}/inventory/{productId}` que devuelve `InventoryItemResponse` joined con `pharmacy_name/slug/district/address`. 404 cuando no existe inventory para ese par.

### K18 — order→user GetAddress hits non-existent endpoint (detectado 2026-05-13)

Mismo patrón que K16, lado user-service. `order-service/internal/infrastructure/clients/user_client_impl.go:28` invoca:
```
GET {USER}/api/v1/users/me/addresses/{addressID}
```
**user-service no expone esa ruta como GET.** Solo existen:
- `GET /api/v1/users/me/addresses` (list)
- `PUT /api/v1/users/me/addresses/{id}` (update, auth)
- `DELETE /api/v1/users/me/addresses/{id}` (delete, auth)

**Síntoma:** user-service responde `405 Method Not Allowed`. order trata non-200 como error → response `400 BUS_028 "Dirección de entrega no encontrada"`. Bloquea `POST /orders/checkout` para CUALQUIER address válida.

**Reproducción:** STEP 9 de `validate-e2e-local.sh` (con K16 ya fixeado).

**Scope:** afecta checkout. Lado código (no IaC).

**Fix recomendado (opción A, simétrico al K16):** agregar handler `GetAddressByID` en user-service + ruta `GET /api/v1/users/me/addresses/{id}`. Mantiene la nomenclatura REST y reusa la repo existente. ~6 archivos (query, handler, controller method, route, main wiring, response DTO reuse).

**Opción alternativa B:** cambiar order's `UserClient.GetAddress` para hacer `GET /addresses` (list) y filtrar por ID client-side. Una respuesta más grande pero sin tocar user-service. Aceptable mientras un usuario tenga pocas direcciones.

### K17 — order no enriquece snapshot con catalog (detectado 2026-05-13)

`add_cart_item_handler.go` hoy lee `ProductName` desde el response de pharmacy. Pero pharmacy **no es fuente de verdad** del nombre del producto — esa data vive en catalog. El endpoint `GET /pharmacies/{id}/inventory/{productId}` devuelve `product_name=""` intencionalmente (pharmacy no resuelve nombres cross-service).

**Síntoma:** `cart_items.product_name` queda vacío. La columna es `NOT NULL DEFAULT ''`, por lo que no rompe el flujo. Pero el snapshot que se guarda en `order_items.product_snapshot` JSONB también queda con `ProductName=""`, lo cual degrada el UX cuando el usuario revisa una orden histórica.

**Fix futuro:** en `AddCartItemHandler.Handle` (y el equivalente en `CheckoutHandler` si reutiliza el flujo), llamar a `catalogClient.GetProduct(ctx, productID)` para enriquecer `product_name`, y solo entonces persistir el `CartItem`. La env var `CATALOG_SERVICE_URL` ya está cableada (K15) — falta el call site.

**Tech debt:** resolver antes de la primera demo con UX que muestre historial de órdenes, o antes de Sprint 4 si Alert Service va a leer `product_snapshot` para construir notificaciones legibles.

**Estado:** ✅ resuelto 2026-06-01. `AddCartItemHandler` recibe `catalogClient` (DI en `main.go`) y llama `GetProduct(ctx, productID)` después de validar inventario. El `Name` de catalog gana sobre `inventoryItem.ProductName`. Si catalog está caído se loggea WARN y se cae al valor de inventory (graceful degradation, no bloquea el cart-add). Backfill defensivo en el branch de update: si el item existente tenía `ProductName=""`, se llena en este pass. `CheckoutHandler` no necesitó cambios — lee `cartItem.ProductName` y ya viene enriquecido, así que el `order_items.product_snapshot` JSONB hereda el nombre correcto.

### K19 — PAYMENT_COMPLETED / PAYMENT_FAILED no se publicaban (detectado 2026-05-13)

`CheckoutHandler` publicaba `ORDER_CREATED` tras un pago exitoso pero **no** los eventos `PAYMENT_COMPLETED` ni `PAYMENT_FAILED`, aunque ambos estaban definidos en `internal/domain/events/order_events.go` y documentados en este archivo. Consumers downstream (Alert/Notification service en Sprint 4, Analytics futuro) quedaban ciegos al resultado del pago — solo veían la orden creada, sin saber si el cobro pasó o falló.

**Síntoma:** ninguno en runtime; degradación silenciosa de la spec de eventos. Detectado durante el e2e del 2026-05-13.

**Estado:** ✅ resuelto 2026-06-01. En `CheckoutHandler.Handle`:
- Rama success: tras `paymentResult.Success`, se publica `PAYMENT_COMPLETED` por cada orden creada (paralelo a `ORDER_CREATED`), con `transaction_id` y `payment_method` en `Metadata`.
- Rama failure: en el loop que marca `payment_status=failed`, se publica `PAYMENT_FAILED` por orden, con `transaction_id` (si lo hay) y `reason` (el `paymentResult.Message`) en `Metadata`.
- Mismo patrón fire-and-forget que el resto del handler (`go func(o *entities.Order)` + `context.Background()`).
- Una transacción de pago cubre N órdenes (multi-farmacia) → se emiten N eventos con el mismo `transaction_id` para que el consumer pueda correlacionar.

### Tech debt — request logging middleware

order-service no loggea requests HTTP por default (solo logs estructurados desde handlers). Diagnosticar issues cross-service (como K16) requirió curl directo a downstream para inferir el flujo. Agregar middleware Zap-based que loggee method+path+status+latency por request mejoraría observabilidad local. Pharmacy-service ya tiene `middleware.Logger` de chi — replicar el patrón.

## Coding Conventions

- **CQRS**: Commands for writes, Queries for reads
- **ApiResponse[T] standard** — ALL endpoints (including health check) return `ApiResponse[T]` via `common.OkResponse()`, `common.CreatedResponse()`, etc. No raw JSON responses allowed.
- **Swagger annotations** on every controller endpoint
- **Structured logging** with `zap.Logger`
- **Spanish messages** for user-facing content (meta fields: `datos`, `mensajes`, `resultado`, `idTransaccion`)
- **Repository pattern** — interfaces in domain, implementations in infrastructure
- **Compile-time interface checks** — `var _ Interface = (*Impl)(nil)`
- **JWT validation only** — this service does NOT generate tokens
- **Pharmacy owner middleware** — order management requires `role == "pharmacy_owner"`
- **Admin middleware** — stats/list-all require `role == "admin"`
- **Fire-and-forget events** — SQS publishing in goroutines
- **Product snapshots** — order items store product data at time of purchase
- **Address snapshots** — orders store delivery address at time of purchase
- **Payment abstraction** — PaymentService interface for easy provider swap
- **respondJSON helper** — controller method that extracts HTTP status from `ApiResponse[T].GetHttpStatus()` and serializes
