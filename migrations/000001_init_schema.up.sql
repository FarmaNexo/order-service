CREATE SCHEMA IF NOT EXISTS orders;

-- Carrito de compras (sesión activa del usuario)
CREATE TABLE orders.cart_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    product_id UUID NOT NULL,
    pharmacy_id UUID NOT NULL,
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(10, 2) NOT NULL,
    product_name VARCHAR(255) NOT NULL DEFAULT '',
    pharmacy_name VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT valid_quantity CHECK (quantity > 0),
    CONSTRAINT valid_price CHECK (unit_price >= 0),
    UNIQUE(user_id, product_id, pharmacy_id)
);

CREATE INDEX idx_cart_items_user_id ON orders.cart_items(user_id);
CREATE INDEX idx_cart_items_product_id ON orders.cart_items(product_id);

-- Órdenes (una por farmacia)
CREATE TABLE orders.orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number VARCHAR(50) NOT NULL UNIQUE,
    user_id UUID NOT NULL,
    pharmacy_id UUID NOT NULL,

    -- Totales
    subtotal DECIMAL(10, 2) NOT NULL,
    delivery_fee DECIMAL(10, 2) NOT NULL DEFAULT 0,
    total DECIMAL(10, 2) NOT NULL,

    -- Delivery
    delivery_method VARCHAR(20) NOT NULL,
    delivery_address_id UUID,
    delivery_address_snapshot JSONB,

    -- Pago
    payment_method VARCHAR(20) NOT NULL,
    payment_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    payment_transaction_id VARCHAR(100),

    -- Estado
    status VARCHAR(20) NOT NULL DEFAULT 'pending_payment',

    -- Info adicional
    notes TEXT,
    cancellation_reason TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    cancelled_at TIMESTAMP,

    CONSTRAINT valid_delivery_method CHECK (delivery_method IN ('delivery', 'pickup')),
    CONSTRAINT valid_payment_method CHECK (payment_method IN ('card', 'cash', 'yape', 'plin')),
    CONSTRAINT valid_payment_status CHECK (payment_status IN ('pending', 'paid', 'failed', 'refunded')),
    CONSTRAINT valid_status CHECK (status IN ('pending_payment', 'confirmed', 'preparing', 'ready', 'in_delivery', 'completed', 'cancelled'))
);

CREATE INDEX idx_orders_user_id ON orders.orders(user_id);
CREATE INDEX idx_orders_pharmacy_id ON orders.orders(pharmacy_id);
CREATE INDEX idx_orders_status ON orders.orders(status);
CREATE INDEX idx_orders_created_at ON orders.orders(created_at DESC);
CREATE INDEX idx_orders_order_number ON orders.orders(order_number);

-- Items de cada orden (snapshot al momento de compra)
CREATE TABLE orders.order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders.orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    product_snapshot JSONB NOT NULL,
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(10, 2) NOT NULL,
    subtotal DECIMAL(10, 2) NOT NULL,

    CONSTRAINT valid_quantity CHECK (quantity > 0),
    CONSTRAINT valid_prices CHECK (unit_price >= 0 AND subtotal >= 0)
);

CREATE INDEX idx_order_items_order_id ON orders.order_items(order_id);
CREATE INDEX idx_order_items_product_id ON orders.order_items(product_id);

-- Historial de cambios de estado
CREATE TABLE orders.order_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders.orders(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL,
    notes TEXT,
    changed_by_user_id UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_status_history_order_id ON orders.order_status_history(order_id);
CREATE INDEX idx_status_history_created_at ON orders.order_status_history(created_at DESC);

-- Secuencia para order_number
CREATE SEQUENCE orders.order_number_seq START WITH 1 INCREMENT BY 1;
