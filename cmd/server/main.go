package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/config"
	"github.com/greenpos/backend/internal/database"
	"github.com/greenpos/backend/internal/graph"
	"github.com/greenpos/backend/internal/middleware"
	"github.com/greenpos/backend/internal/repository"
	"github.com/greenpos/backend/internal/service"
)

func main() {
	// Initialize structured logger
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database (optional - app works without DB for development)
	var db *database.PostgresDB
	if cfg.DatabaseURL != "" {
		db, err = database.NewPostgresDB(cfg, logger)
		if err != nil {
			logger.Warn("Failed to connect to database, using in-memory storage",
				slog.String("error", err.Error()))
		} else {
			defer db.Close()

			// Run migrations
			ctx := context.Background()
			if err := database.RunMigrations(ctx, db.Pool(), logger); err != nil {
				log.Fatalf("Failed to run migrations: %v", err)
			}
		}
	}

	// Initialize repositories (memory or database-backed)
	var repos *repository.Repositories
	if db != nil {
		repos = repository.NewDBRepositories(db.DB)
	} else {
		repos = repository.NewMemoryRepositories()
	}

	// Seed data if using in-memory storage
	if db == nil {
		ctx := context.Background()
		if err := repos.Seed(ctx); err != nil {
			logger.Warn("Failed to seed data", slog.String("error", err.Error()))
		}
	}

	// Create service configuration
	serviceCfg := &service.ServiceConfig{
		JWTSecret:             cfg.JWTSecret,
		JWTExpiryHours:        cfg.JWTExpiryHours,
		RefreshTokenExpiryDays: cfg.RefreshTokenExpiryDays,
		BCryptCost:            cfg.BCryptCost,
		TaxRate:               0.16, // Default 16% tax rate
	}

	// Initialize services
	services := service.NewServices(repos, serviceCfg, logger)

	// Initialize GraphQL resolver
	resolvers := graph.NewResolvers(services, cfg.JWTSecret)

	// GraphQL server
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolvers}))

	// Initialize middleware
	validator := middleware.NewValidator()
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRequestsPerMinute, time.Minute)
	defer rateLimiter.Stop()

	corsConfig := middleware.DefaultCORSConfig()
	corsConfig.AllowedOrigins = cfg.CORSAllowedOrigins

	// Create HTTP handlers
	graphqlHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/graphql" && r.Method == "POST" {
			srv.ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	})

	playgroundHandler := playground.Handler("GreenPOS GraphQL", "/graphql")

	// Build router with middleware
	router := http.NewServeMux()
	router.Handle("/graphql", graphqlHandler)
	router.Handle("/graphql/playground", playgroundHandler)
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if db != nil {
			if err := db.Ping(r.Context()); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(`{"status":"unhealthy","database":"down"}`))
				return
			}
		}
		w.Write([]byte(`{"status":"healthy","database":"up"}`))
	})

	// Webhook handlers for payment providers
	router.HandleFunc("/webhooks/stripe", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		payload, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("Failed to read Stripe webhook body", slog.String("error", err.Error()))
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		signature := r.Header.Get("Stripe-Signature")

		if err := services.Stripe.HandleWebhook(r.Context(), payload, signature); err != nil {
			logger.Error("Stripe webhook error", slog.String("error", err.Error()))
			http.Error(w, "Webhook error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"received":true}`))
	})

	router.HandleFunc("/webhooks/paypal", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var event map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			logger.Error("Failed to parse PayPal webhook body", slog.String("error", err.Error()))
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if err := services.PayPal.HandleWebhook(r.Context(), event); err != nil {
			logger.Error("PayPal webhook error", slog.String("error", err.Error()))
			http.Error(w, "Webhook error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"received":true}`))
	})

	// Apply middleware chain
	finalHandler := middleware.RecoveryMiddleware(logger)
	finalHandler = middleware.LoggingMiddleware(logger)(finalHandler)
	finalHandler = middleware.SecurityHeadersMiddleware()(finalHandler)
	finalHandler = middleware.CORSMiddleware(corsConfig)(finalHandler)
	finalHandler = middleware.RateLimitMiddleware(rateLimiter, logger)(finalHandler)
	finalHandler = middleware.RequestIDMiddleware()(finalHandler)

	// Wrap with middleware
	handlerWithMiddleware := middleware.TimeoutMiddleware(30 * time.Second)(finalHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		router.ServeHTTP(w, r)
	})))

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handlerWithMiddleware,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("starting server",
			slog.String("port", cfg.Port),
			slog.String("environment", cfg.Environment),
			slog.String("playground", fmt.Sprintf("http://localhost:%s/graphql/playground", cfg.Port)),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", slog.String("error", err.Error()))
	}

	// Close database connection
	if db != nil {
		if err := db.Close(); err != nil {
			logger.Error("database close error", slog.String("error", err.Error()))
		}
	}

	logger.Info("server stopped")
}

// GenerateToken generates a JWT token for testing
func GenerateToken(userID, tenantID, branchID uuid.UUID, role string, secret string) (string, error) {
	config := &middleware.JWTConfig{
		Secret:               []byte(secret),
		ExpiryHours:          24,
		RefreshTokenExpiryDays: 7,
	}
	authSvc := middleware.NewAuthService(config, nil)

	tokenUser := &middleware.TokenUser{
		ID:       userID,
		TenantID:  tenantID,
		BranchID: branchID,
		Email:    "test@example.com",
		Role:     role,
	}

	accessToken, _, err := authSvc.GenerateTokenPair(tokenUser)
	return accessToken, err
}
