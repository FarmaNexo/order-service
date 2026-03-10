# FarmaNexo Order Service

Servicio de gestión de órdenes y carrito de compras para la plataforma FarmaNexo.

## Características

- **Carrito multi-farmacia**: Agrega productos de diferentes farmacias en un solo carrito
- **Checkout inteligente**: Genera órdenes separadas por farmacia automáticamente
- **Gestión de pagos**: Integración con pasarela de pagos (mock para desarrollo)
- **Gestión de estados**: Flujo completo de estados de orden con validación de transiciones
- **Panel de farmacia**: Endpoints para que farmacias gestionen sus pedidos
- **Panel admin**: Estadísticas y vista global de órdenes

## Stack Tecnológico

- **Go 1.23+** con Chi router
- **PostgreSQL** con GORM
- **Redis** para cache
- **SQS** para eventos asíncronos
- **JWT** para autenticación (validación compartida)
- **Clean Architecture** + CQRS + MediatR

## Inicio Rápido

```bash
# Instalar dependencias
make install

# Aplicar migraciones
make migrate-up

# Ejecutar en modo local
make dev
```

El servicio estará disponible en `http://localhost:4011`.
Swagger UI: `http://localhost:4011/swagger/index.html`

## Estructura de Respuesta Estándar

Todos los endpoints (incluyendo health check) retornan la estructura `ApiResponse[T]`:

```json
{
  "meta": {
    "mensajes": [
      {
        "codigo": "SUC_001",
        "mensaje": "Operación exitosa",
        "tipo": "information"
      }
    ],
    "idTransaccion": "550e8400-e29b-41d4-a716-446655440000",
    "resultado": true,
    "timestamp": "20260223 153045"
  },
  "datos": { ... }
}
```

## Endpoints

### Health Check (sin autenticación)

```bash
# Verificar estado del servicio
curl http://localhost:4011/health
```

Respuesta:
```json
{
  "meta": { "resultado": true, "mensajes": [...] },
  "datos": {
    "status": "healthy",
    "service": "order-service",
    "version": "1.0.0"
  }
}
```

### Carrito (requiere JWT)

```bash
# Ver carrito
curl http://localhost:4011/api/v1/cart -H "Authorization: Bearer {token}"

# Agregar producto
curl -X POST http://localhost:4011/api/v1/cart/items \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{"product_id": "uuid", "pharmacy_id": "uuid", "quantity": 2}'

# Actualizar cantidad
curl -X PUT http://localhost:4011/api/v1/cart/items/{item_id} \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{"quantity": 3}'

# Eliminar item
curl -X DELETE http://localhost:4011/api/v1/cart/items/{item_id} \
  -H "Authorization: Bearer {token}"

# Vaciar carrito
curl -X DELETE http://localhost:4011/api/v1/cart \
  -H "Authorization: Bearer {token}"
```

### Checkout y Órdenes

```bash
# Checkout (crear órdenes + pagar)
curl -X POST http://localhost:4011/api/v1/orders/checkout \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "delivery_method": "delivery",
    "delivery_address_id": "uuid",
    "payment_method": "yape",
    "notes": "Entregar en recepción"
  }'

# Listar mis órdenes
curl "http://localhost:4011/api/v1/orders?status=confirmed&page=1&limit=10" \
  -H "Authorization: Bearer {token}"

# Ver detalle de orden
curl http://localhost:4011/api/v1/orders/{order_id} \
  -H "Authorization: Bearer {token}"

# Cancelar orden
curl -X POST http://localhost:4011/api/v1/orders/{order_id}/cancel \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{"reason": "Cambié de opinión"}'
```

### Farmacia (requiere rol pharmacy_owner)

```bash
# Ver órdenes de mi farmacia
curl "http://localhost:4011/api/v1/pharmacy/orders?status=confirmed" \
  -H "Authorization: Bearer {token}"

# Actualizar estado de orden
curl -X PUT http://localhost:4011/api/v1/pharmacy/orders/{order_id}/status \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{"status": "preparing", "notes": "Preparando pedido"}'
```

### Admin (requiere rol admin)

```bash
# Ver todas las órdenes
curl "http://localhost:4011/api/v1/admin/orders?status=confirmed&page=1" \
  -H "Authorization: Bearer {token}"

# Estadísticas
curl "http://localhost:4011/api/v1/admin/orders/stats?date_from=2026-01-01&date_to=2026-12-31" \
  -H "Authorization: Bearer {token}"
```

## Flujo de Checkout

```
1. Usuario agrega productos al carrito (de múltiples farmacias)
2. Al hacer checkout:
   a. Se valida el carrito (no vacío, stock disponible)
   b. Se valida la dirección de entrega (si es delivery)
   c. Se agrupan items por farmacia
   d. Se crea UNA orden por cada farmacia
   e. Se genera order_number único (ORD-2026-000001)
   f. Se calcula delivery_fee por orden
   g. Se procesa el pago (total de todas las órdenes)
   h. Si exitoso: confirmar órdenes, vaciar carrito, publicar eventos
   i. Si falla: mantener carrito, marcar error
3. Cada farmacia recibe notificación de su orden via SQS
```

## Estados de Orden

```
pending_payment → confirmed → preparing → ready → in_delivery → completed
       ↓              ↓
    cancelled      cancelled
```

## Base de Datos

- **Database**: `order_db`
- **Schema**: `orders`
- **Tablas**: `cart_items`, `orders`, `order_items`, `order_status_history`

## Variables de Entorno

| Variable | Descripción | Default (local) |
|---|---|---|
| ENV | Ambiente | local |
| DB_HOST | Host PostgreSQL | localhost |
| DB_PASSWORD | Password PostgreSQL | admin |
| JWT_SECRET | Secret JWT compartido | dev-super-secret-key... |
| REDIS_HOST | Host Redis | localhost |
| REDIS_PASSWORD | Password Redis | farmanexo2026 |

## Integración con Pasarela de Pago

El servicio usa una interfaz `PaymentService` que permite cambiar fácilmente el proveedor:

```go
type PaymentService interface {
    ProcessPayment(ctx context.Context, request PaymentRequest) (*PaymentResult, error)
    RefundPayment(ctx context.Context, transactionID string, amount float64) (*PaymentResult, error)
}
```

En desarrollo se usa `MockPaymentService` con 80% de éxito. Para producción, implementar con Niubiz, Culqi, MercadoPago, etc.
