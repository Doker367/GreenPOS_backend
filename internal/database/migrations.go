package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Migration represents a single database migration
type Migration struct {
	Version     int
	Description string
	Up          string
	Down        string
}

// migrations holds all database migrations in order
var migrations = []Migration{
	{
		Version:     1,
		Description: "Enable UUID extension and create tenants table",
		Up: `CREATE EXTENSION IF NOT EXISTS "pgcrypto"`,
		Down: `DROP EXTENSION IF EXISTS "pgcrypto" CASCADE`,
	},
	{
		Version:     2,
		Description: "Create schema_migrations table",
		Up: `CREATE TABLE IF NOT EXISTS schema_migrations (
			version INT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		Down: `DROP TABLE IF EXISTS schema_migrations`,
	},
	{
		Version:     3,
		Description: "Create tenants table",
		Up: `CREATE TABLE IF NOT EXISTS tenants (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(100) UNIQUE NOT NULL,
			settings JSONB DEFAULT '{}',
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		Down: `DROP TABLE IF EXISTS tenants`,
	},
	{
		Version:     4,
		Description: "Create branches table",
		Up: `CREATE TABLE IF NOT EXISTS branches (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			address TEXT,
			phone VARCHAR(50),
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		Down: `DROP TABLE IF EXISTS branches`,
	},
	{
		Version:     5,
		Description: "Create users table",
		Up: `CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
			email VARCHAR(255) NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL CHECK (role IN ('OWNER', 'ADMIN', 'MANAGER', 'WAITER', 'KITCHEN')),
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(branch_id, email)
		)`,
		Down: `DROP TABLE IF EXISTS users`,
	},
	{
		Version:     6,
		Description: "Create categories table",
		Up: `CREATE TABLE IF NOT EXISTS categories (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			sort_order INT DEFAULT 0,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		Down: `DROP TABLE IF EXISTS categories`,
	},
	{
		Version:     7,
		Description: "Create products table",
		Up: `CREATE TABLE IF NOT EXISTS products (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
			category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			price DECIMAL(10,2) NOT NULL,
			image_url TEXT,
			is_available BOOLEAN DEFAULT true,
			is_featured BOOLEAN DEFAULT false,
			preparation_time INT DEFAULT 15,
			allergens TEXT[] DEFAULT '{}',
			rating DECIMAL(3,2) DEFAULT 0,
			review_count INT DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		Down: `DROP TABLE IF EXISTS products`,
	},
	{
		Version:     8,
		Description: "Create tables table (restaurant tables)",
		Up: `CREATE TABLE IF NOT EXISTS tables (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
			number VARCHAR(20) NOT NULL,
			capacity INT DEFAULT 4,
			status VARCHAR(50) DEFAULT 'AVAILABLE' CHECK (status IN ('AVAILABLE', 'OCCUPIED', 'RESERVED')),
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(branch_id, number)
		)`,
		Down: `DROP TABLE IF EXISTS tables`,
	},
	{
		Version:     9,
		Description: "Create orders table",
		Up: `CREATE TABLE IF NOT EXISTS orders (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
			table_id UUID REFERENCES tables(id) ON DELETE SET NULL,
			user_id UUID NOT NULL REFERENCES users(id),
			customer_name VARCHAR(255),
			customer_phone VARCHAR(50),
			status VARCHAR(50) DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'ACCEPTED', 'PREPARING', 'READY', 'DELIVERED', 'CANCELLED', 'PAID')),
			subtotal DECIMAL(10,2) DEFAULT 0,
			tax DECIMAL(10,2) DEFAULT 0,
			discount DECIMAL(10,2) DEFAULT 0,
			total DECIMAL(10,2) DEFAULT 0,
			notes TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		Down: `DROP TABLE IF EXISTS orders`,
	},
	{
		Version:     10,
		Description: "Create order_items table",
		Up: `CREATE TABLE IF NOT EXISTS order_items (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
			quantity INT NOT NULL DEFAULT 1,
			unit_price DECIMAL(10,2) NOT NULL,
			total_price DECIMAL(10,2) NOT NULL,
			notes TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		Down: `DROP TABLE IF EXISTS order_items`,
	},
	{
		Version:     11,
		Description: "Create reservations table",
		Up: `CREATE TABLE IF NOT EXISTS reservations (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
			table_id UUID REFERENCES tables(id) ON DELETE SET NULL,
			customer_name VARCHAR(255) NOT NULL,
			customer_phone VARCHAR(50),
			customer_email VARCHAR(255),
			guest_count INT DEFAULT 1,
			reservation_date DATE NOT NULL,
			reservation_time TIME NOT NULL,
			status VARCHAR(50) DEFAULT 'CONFIRMED' CHECK (status IN ('PENDING', 'CONFIRMED', 'CANCELLED', 'COMPLETED')),
			notes TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		Down: `DROP TABLE IF EXISTS reservations`,
	},
	{
		Version:     12,
		Description: "Create refresh_tokens table for JWT refresh token support",
		Up: `CREATE TABLE IF NOT EXISTS refresh_tokens (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash VARCHAR(255) NOT NULL UNIQUE,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			revoked_at TIMESTAMPTZ
		)`,
		Down: `DROP TABLE IF EXISTS refresh_tokens`,
	},
	{
		Version:     13,
		Description: "Create indexes for performance",
		Up: `CREATE INDEX IF NOT EXISTS idx_branches_tenant_id ON branches(tenant_id);
			CREATE INDEX IF NOT EXISTS idx_users_branch_id ON users(branch_id);
			CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
			CREATE INDEX IF NOT EXISTS idx_categories_branch_id ON categories(branch_id);
			CREATE INDEX IF NOT EXISTS idx_products_branch_id ON products(branch_id);
			CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
			CREATE INDEX IF NOT EXISTS idx_tables_branch_id ON tables(branch_id);
			CREATE INDEX IF NOT EXISTS idx_orders_branch_id ON orders(branch_id);
			CREATE INDEX IF NOT EXISTS idx_orders_table_id ON orders(table_id);
			CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
			CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
			CREATE INDEX IF NOT EXISTS idx_reservations_branch_id ON reservations(branch_id);
			CREATE INDEX IF NOT EXISTS idx_reservations_date ON reservations(reservation_date);`,
		Down: `DROP INDEX IF EXISTS idx_reservations_date;
			DROP INDEX IF EXISTS idx_reservations_branch_id;
			DROP INDEX IF EXISTS idx_order_items_order_id;
			DROP INDEX IF EXISTS idx_orders_status;
			DROP INDEX IF EXISTS idx_orders_table_id;
			DROP INDEX IF EXISTS idx_orders_branch_id;
			DROP INDEX IF EXISTS idx_tables_branch_id;
			DROP INDEX IF EXISTS idx_products_category_id;
			DROP INDEX IF EXISTS idx_products_branch_id;
			DROP INDEX IF EXISTS idx_categories_branch_id;
			DROP INDEX IF EXISTS idx_users_email;
			DROP INDEX IF EXISTS idx_users_branch_id;
			DROP INDEX IF EXISTS idx_branches_tenant_id;`,
	},
}

// RunMigrations executes all pending database migrations
func RunMigrations(ctx context.Context, pool *pgxpool.Pool, log *slog.Logger) error {
	if log != nil {
		log.Info("starting database migrations")
	}

	// Ensure migrations table exists
	_, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version INT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		applied_at TIMESTAMPTZ DEFAULT NOW()
	)`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get current version
	var currentVersion int
	err = pool.QueryRow(ctx, `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&currentVersion)
	if err != nil {
		currentVersion = 0
	}

	if log != nil {
		log.Info("current migration version", slog.Int("version", currentVersion))
	}

	// Apply pending migrations
	for i := currentVersion; i < len(migrations); i++ {
		m := migrations[i]
		start := time.Now()

		// Execute migration
		if _, err := pool.Exec(ctx, m.Up); err != nil {
			return fmt.Errorf("migration v%d (%s) failed: %w", m.Version, m.Description, err)
		}

		// Record migration
		_, err := pool.Exec(ctx,
			`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`,
			m.Version, m.Description)
		if err != nil {
			return fmt.Errorf("failed to record migration v%d: %w", m.Version, err)
		}

		if log != nil {
			log.Info("applied migration",
				slog.Int("version", m.Version),
				slog.String("description", m.Description),
				slog.Duration("duration", time.Since(start)),
			)
		}
	}

	if log != nil {
		log.Info("migrations completed", slog.Int("total_migrations", len(migrations)))
	}

	return nil
}

// RollbackMigration rolls back the most recent migration
func RollbackMigration(ctx context.Context, pool *pgxpool.Pool, log *slog.Logger) error {
	// Get current version
	var currentVersion int
	err := pool.QueryRow(ctx, `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if currentVersion == 0 {
		return fmt.Errorf("no migrations to rollback")
	}

	// Find the migration to rollback
	var migration Migration
	found := false
	for _, m := range migrations {
		if m.Version == currentVersion {
			migration = m
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("migration v%d not found", currentVersion)
	}

	// Execute rollback
	if _, err := pool.Exec(ctx, migration.Down); err != nil {
		return fmt.Errorf("rollback v%d failed: %w", currentVersion, err)
	}

	// Remove migration record
	_, err = pool.Exec(ctx, `DELETE FROM schema_migrations WHERE version = $1`, currentVersion)
	if err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	if log != nil {
		log.Info("rolled back migration",
			slog.Int("version", currentVersion),
			slog.String("description", migration.Description),
		)
	}

	return nil
}

// GetMigrationVersion returns the current migration version
func GetMigrationVersion(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	var version int
	err := pool.QueryRow(ctx, `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}
