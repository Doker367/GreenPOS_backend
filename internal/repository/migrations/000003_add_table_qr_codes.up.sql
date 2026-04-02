-- Migration: Add table_qr_codes table
-- Supports QR menu functionality where each table has a unique QR code

CREATE TABLE IF NOT EXISTS table_qr_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_id UUID NOT NULL REFERENCES pos_tables(id) ON DELETE CASCADE,
    qr_token VARCHAR(255) UNIQUE NOT NULL,
    access_url VARCHAR(500) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for faster lookups by token
CREATE INDEX IF NOT EXISTS idx_table_qr_codes_token ON table_qr_codes(qr_token);
CREATE INDEX IF NOT EXISTS idx_table_qr_codes_table_id ON table_qr_codes(table_id);
CREATE INDEX IF NOT EXISTS idx_table_qr_codes_active ON table_qr_codes(is_active) WHERE is_active = true;
