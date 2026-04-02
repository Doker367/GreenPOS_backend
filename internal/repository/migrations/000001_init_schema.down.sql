-- Migration: 000001_init_schema.down.sql
-- Desc: Rollback complete GreenPOS schema

DROP INDEX IF EXISTS idx_invoice_items_invoice;
DROP INDEX IF EXISTS idx_invoices_order;
DROP INDEX IF EXISTS idx_invoices_branch;
DROP INDEX IF EXISTS idx_reservations_date;
DROP INDEX IF EXISTS idx_reservations_branch;
DROP INDEX IF EXISTS idx_order_items_order;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_table;
DROP INDEX IF EXISTS idx_orders_branch;
DROP INDEX IF EXISTS idx_pos_tables_branch;
DROP INDEX IF EXISTS idx_products_category;
DROP INDEX IF EXISTS idx_products_branch;
DROP INDEX IF EXISTS idx_categories_branch;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_branch;
DROP INDEX IF EXISTS idx_branches_tenant;

DROP TABLE IF EXISTS tenant_fiscal;
DROP TABLE IF EXISTS invoice_items;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS reservations;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS pos_tables;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS branches;
DROP TABLE IF EXISTS tenants;

DROP EXTENSION IF EXISTS "pgcrypto";
