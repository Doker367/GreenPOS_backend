package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres(databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

func RunMigrations(pool *pgxpool.Pool) error {
	migrations := []string{
		`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`,
		
		`CREATE TABLE IF NOT EXISTS tenants (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(100) UNIQUE NOT NULL,
			settings JSONB DEFAULT '{}',
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS branches (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			address TEXT,
			phone VARCHAR(50),
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS users (
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

		`CREATE TABLE IF NOT EXISTS categories (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			sort_order INT DEFAULT 0,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS products (
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

		`CREATE TABLE IF NOT EXISTS tables (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
			number VARCHAR(20) NOT NULL,
			capacity INT DEFAULT 4,
			status VARCHAR(50) DEFAULT 'AVAILABLE' CHECK (status IN ('AVAILABLE', 'OCCUPIED', 'RESERVED')),
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(branch_id, number)
		)`,

		`CREATE TABLE IF NOT EXISTS orders (
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

		`CREATE TABLE IF NOT EXISTS order_items (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
			quantity INT NOT NULL DEFAULT 1,
			unit_price DECIMAL(10,2) NOT NULL,
			total_price DECIMAL(10,2) NOT NULL,
			notes TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS reservations (
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
	}

	for _, migration := range migrations {
		if _, err := pool.Exec(context.Background(), migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}
