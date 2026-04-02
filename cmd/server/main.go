package main

import (
	"log"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/config"
	"github.com/greenpos/backend/internal/database"
	"github.com/greenpos/backend/internal/graph"
	"github.com/greenpos/backend/internal/repository"
	"github.com/greenpos/backend/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize layers
	repos := repository.NewRepositories(db)
	services := service.NewServices(repos, cfg.JWTSecret)
	resolvers := graph.NewResolvers(services, cfg.JWTSecret)

	// GraphQL server
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolvers}))

	// CORS middleware
	coreHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/graphql" && r.Method == "POST" {
			srv.ServeHTTP(w, r)
			return
		}

		if r.URL.Path == "/health" {
			w.Write([]byte(`{"status":"ok"}`))
			return
		}

		if r.URL.Path == "/graphql/playground" {
			playground.Handler("GreenPOS GraphQL", "/graphql").ServeHTTP(w, r)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	// Start server
	log.Printf("🚀 Server starting on port %s", cfg.Port)
	log.Printf("📊 GraphQL Playground: http://localhost:%s/graphql/playground", cfg.Port)

	if err := http.ListenAndServe(":"+cfg.Port, http.HandlerFunc(coreHandler)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// GenerateToken generates a JWT token for testing
func GenerateToken(userID, tenantID, branchID uuid.UUID, role string, secret string) (string, error) {
	claims := jwt.MapClaims{
		"sub":    userID.String(),
		"tenant": tenantID.String(),
		"branch": branchID.String(),
		"role":   role,
		"exp":    jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
