-- Migration: 000002_add_payments.up.sql
-- Desc: Add payments table for Stripe and PayPal integration

CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    amount DECIMAL(10,2) NOT NULL,
    method VARCHAR(20) NOT NULL CHECK (method IN ('cash', 'card', 'transfer')),
    provider VARCHAR(20) CHECK (provider IN ('stripe', 'paypal') OR provider IS NULL),
    provider_payment_id VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed', 'refunded')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for faster lookups by order
CREATE INDEX idx_payments_order ON payments(order_id);

-- Index for status queries
CREATE INDEX idx_payments_status ON payments(status);

-- Index for provider lookups
CREATE INDEX idx_payments_provider ON payments(provider);
