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
