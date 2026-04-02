# GreenPOS Backend

A GraphQL-based Point of Sale system for restaurants with multi-tenancy support.

## Features

- **Multi-tenant Architecture**: Each restaurant group (tenant) can have multiple branches
- **Branch Management**: Each branch operates independently with its own users, products, orders, and reservations
- **Role-based Access**: OWNER, ADMIN, MANAGER, WAITER, KITCHEN roles
- **Order Management**: Full order lifecycle from PENDING to PAID
- **Table Management**: Track table status (AVAILABLE, OCCUPIED, RESERVED)
- **Reservation System**: Manage customer reservations with date/time tracking
- **Category & Product Management**: Organize products with categories, featured items, allergens

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      GraphQL API                             │
│                  (internal/graph/)                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │  Resolvers  │  │  Auth MW    │  │  Context Helpers    │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└───────────────────────────┬─────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────┐
│                     Service Layer                            │
│                  (internal/service/)                        │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Auth │ Tenant │ Branch │ User │ Category │ Product     │ │
│  │ Table │ Order │ Reservation                             │ │
│  └────────────────────────────────────────────────────────┘ │
└───────────────────────────┬─────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────┐
│                   Repository Layer                           │
│                 (internal/repository/)                       │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Tenants │ Branches │ Users │ Categories │ Products    │ │
│  │ Tables │ Orders │ Reservations                            │ │
│  └────────────────────────────────────────────────────────┘ │
└───────────────────────────┬─────────────────────────────────┘
                            │
                    ┌───────▼───────┐
                    │   PostgreSQL   │
                    └───────────────┘
```

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Docker (optional)

### Using Docker

```bash
# Build and run
docker build -t greenpos-backend .
docker run -p 8080:8080 \
  -e DATABASE_URL=postgres://user:pass@host:5432/greenpos \
  -e JWT_SECRET=your-secret-key \
  greenpos-backend
```

### Local Development

```bash
# Install dependencies
go mod download

# Set environment variables
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/greenpos?sslmode=disable
export JWT_SECRET=your-secret-key-change-in-production
export PORT=8080

# Run database migrations (requires migrate tool)
migrate -path internal/repository/migrations -database "$DATABASE_URL" up

# Start the server
go run cmd/server/main.go
```

### GraphQL Playground

Once running, access the interactive playground at:
```
http://localhost:8080/
```

## Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `DATABASE_URL` | PostgreSQL connection string | Yes | - |
| `JWT_SECRET` | Secret key for JWT signing | Yes | - |
| `PORT` | Server port | No | `8080` |
| `CORS_ALLOWED_ORIGINS` | Allowed CORS origins (comma-separated) | No | `*` |

## API Reference

### Authentication

All mutations require a `Authorization: Bearer <token>` header except `login` and `createTenant`.

### Queries

#### Get Current User
```graphql
query {
  me {
    id
    email
    name
    role
    branch {
      id
      name
    }
  }
}
```

#### List Branches
```graphql
query {
  branches {
    id
    name
    address
  }
}
```

#### Get Products
```graphql
query {
  products(branchId: "uuid-here") {
    id
    name
    price
    category {
      name
    }
  }
}
```

#### List Orders
```graphql
query {
  orders(branchId: "uuid-here", status: PENDING, limit: 10) {
    id
    status
    total
    items {
      product { name }
      quantity
    }
  }
}
```

### Mutations

#### Login
```graphql
mutation {
  login(email: "user@example.com", password: "password", branchId: "uuid-here") {
    token
    user {
      id
      email
      role
    }
  }
}
```

#### Create Tenant
```graphql
mutation {
  createTenant(input: { name: "Restaurant Group", slug: "rg-1" }) {
    id
    name
  }
}
```

#### Create Branch
```graphql
mutation {
  createBranch(input: { tenantId: "uuid", name: "Main Branch", address: "123 Main St" }) {
    id
    name
  }
}
```

#### Create User
```graphql
mutation {
  createUser(input: {
    branchId: "uuid"
    email: "staff@restaurant.com"
    password: "securepassword"
    name: "John Doe"
    role: WAITER
  }) {
    id
    email
    role
  }
}
```

#### Create Product
```graphql
mutation {
  createProduct(input: {
    branchId: "uuid"
    categoryId: "uuid"
    name: "Margherita Pizza"
    description: "Classic tomato and mozzarella"
    price: 12.99
    isAvailable: true
    isFeatured: true
    preparationTime: 15
    allergens: ["gluten", "dairy"]
  }) {
    id
    name
    price
  }
}
```

#### Create Order
```graphql
mutation {
  createOrder(input: {
    branchId: "uuid"
    tableId: "uuid"
    userId: "uuid"
    customerName: "Jane Customer"
    items: [
      { productId: "uuid", quantity: 2, notes: "no onions" }
    ]
  }) {
    id
    status
    total
  }
}
```

#### Update Order Status
```graphql
mutation {
  updateOrderStatus(id: "uuid", status: PREPARING) {
    id
    status
  }
}
```

#### Create Reservation
```graphql
mutation {
  createReservation(input: {
    branchId: "uuid"
    tableId: "uuid"
    customerName: "John Smith"
    customerPhone: "+1234567890"
    guestCount: 4
    reservationDate: "2024-12-25"
    reservationTime: "19:00:00"
  }) {
    id
    customerName
    status
  }
}
```

## Folder Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── graph/
│   │   ├── schema.resolvers.go # GraphQL resolvers
│   │   ├── generated.go       # Generated code (gqlgen)
│   │   ├── models_gen.go      # Generated models (gqlgen)
│   │   └── schema.graphql     # GraphQL schema
│   ├── middleware/
│   │   └── auth.go           # JWT authentication middleware
│   ├── model/
│   │   └── models.go         # Domain models
│   ├── repository/
│   │   ├── repository.go     # Repository interfaces and implementations
│   │   └── *.go              # Individual repository files
│   └── service/
│       ├── service.go        # Service layer with business logic
│       └── *.go              # Individual service files
├── schema/
│   └── schema.graphql        # GraphQL schema definition
├── gqlgen.yml               # gqlgen configuration
├── go.mod
├── go.sum
└── README.md
```

## Multi-Tenancy

The system enforces tenant isolation at the resolver level:

1. JWT token contains `userId`, `branchId`, and `tenantId`
2. Every resolver extracts `tenantId` from context
3. Before accessing any resource, the resolver verifies it belongs to the current tenant
4. Cross-tenant access returns `ErrForbidden`

## Order Status Flow

```
PENDING → ACCEPTED → PREPARING → READY → DELIVERED → PAID
    ↓          ↓
  CANCELLED  CANCELLED
```

## Error Handling

The API uses GraphQL error extensions for error codes:

| Code | Description |
|------|-------------|
| `NOT_FOUND` | Resource not found |
| `FORBIDDEN` | Access denied (tenant isolation) |
| `BAD_REQUEST` | Invalid input |
| `INTERNAL_ERROR` | Server error |

## License

MIT
