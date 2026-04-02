package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/greenpos/backend/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgresDB wraps GORM DB with additional functionality
type PostgresDB struct {
	DB     *gorm.DB
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewPostgresDB creates a new PostgreSQL connection using GORM
func NewPostgresDB(cfg *config.Config, log *slog.Logger) (*PostgresDB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	// Configure GORM logger
	gormLogLevel := logger.Warn
	if cfg.IsDevelopment() {
		gormLogLevel = logger.Info
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Open connection with GORM
	gormDB, err := gorm.Open(postgres.Open(cfg.DatabaseURL), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Get underlying pgxpool
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(int(cfg.DBMaxOpenConns))
	sqlDB.SetMaxIdleConns(int(cfg.DBMaxIdleConns))
	sqlDB.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	// Create pgxpool config for advanced features
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL for pool config: %w", err)
	}

	poolConfig.MaxConns = cfg.DBMaxOpenConns
	poolConfig.MinConns = cfg.DBMaxIdleConns
	poolConfig.MaxConnLifetime = cfg.DBConnMaxLifetime

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if log != nil {
		log.Info("database connection established",
			slog.String("environment", cfg.Environment),
			slog.Int("max_open_conns", int(cfg.DBMaxOpenConns)),
			slog.Int("max_idle_conns", int(cfg.DBMaxIdleConns)),
		)
	}

	return &PostgresDB{
		DB:     gormDB,
		pool:   pool,
		logger: log,
	}, nil
}

// Pool returns the underlying pgxpool.Pool for advanced operations
func (db *PostgresDB) Pool() *pgxpool.Pool {
	return db.pool
}

// Close closes the database connection pool
func (db *PostgresDB) Close() error {
	if db.pool != nil {
		db.pool.Close()
	}
	if db.DB != nil {
		sqlDB, err := db.DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
	if db.logger != nil {
		db.logger.Info("database connection closed")
	}
	return nil
}

// Ping verifies database connection is alive
func (db *PostgresDB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

// Stats returns connection pool statistics
func (db *PostgresDB) Stats() *pgxpool.Stat {
	return db.pool.Stat()
}

// AutoMigrate runs GORM automigration for all models
func (db *PostgresDB) AutoMigrate(models ...interface{}) error {
	if db.logger != nil {
		db.logger.Info("running database auto-migration")
	}

	for _, model := range models {
		if err := db.DB.AutoMigrate(model); err != nil {
			return fmt.Errorf("auto-migration failed for %T: %w", model, err)
		}
	}

	if db.logger != nil {
		db.logger.Info("database auto-migration completed")
	}

	return nil
}

// WithContext returns a GORM DB with the given context
func (db *PostgresDB) WithContext(ctx context.Context) *gorm.DB {
	return db.DB.WithContext(ctx)
}

// Transaction executes a function within a database transaction
func (db *PostgresDB) Transaction(fn func(*gorm.DB) error) error {
	return db.DB.Transaction(fn)
}
