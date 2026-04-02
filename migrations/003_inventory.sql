-- Migration: 003_inventory.sql
-- Creates inventory tracking tables

-- inventory_items: tracks stock of ingredients/products
CREATE TABLE IF NOT EXISTS inventory_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    unit VARCHAR(50) NOT NULL, -- kg, litros, piezas, cajas, etc.
    current_stock DECIMAL(10,2) DEFAULT 0,
    min_stock DECIMAL(10,2) DEFAULT 0,
    cost DECIMAL(10,2) DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- inventory_movements: tracks stock changes
CREATE TABLE IF NOT EXISTS inventory_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inventory_item_id UUID NOT NULL REFERENCES inventory_items(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('IN', 'OUT', 'ADJUSTMENT')),
    quantity DECIMAL(10,2) NOT NULL,
    reason VARCHAR(255),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_inventory_items_branch ON inventory_items(branch_id);
CREATE INDEX IF NOT EXISTS idx_inventory_items_active ON inventory_items(branch_id, is_active);
CREATE INDEX IF NOT EXISTS idx_inventory_movements_item ON inventory_movements(inventory_item_id);
CREATE INDEX IF NOT EXISTS idx_inventory_movements_created ON inventory_movements(created_at DESC);

-- Comments
COMMENT ON TABLE inventory_items IS 'Tracks stock levels for ingredients and products';
COMMENT ON TABLE inventory_movements IS 'Audit log of all stock changes';
COMMENT ON COLUMN inventory_items.unit IS 'Measurement unit: kg, litros, piezas, cajas, gramos, ml';
COMMENT ON COLUMN inventory_items.min_stock IS 'Minimum stock level - items below this trigger low-stock alerts';
COMMENT ON COLUMN inventory_movements.type IS 'IN=addition, OUT=removal, ADJUSTMENT=correction';
