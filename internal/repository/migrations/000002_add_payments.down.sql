-- Migration: 000002_add_payments.down.sql
-- Desc: Rollback payments table

DROP INDEX IF EXISTS idx_payments_provider;
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_payments_order;
DROP TABLE IF EXISTS payments;
