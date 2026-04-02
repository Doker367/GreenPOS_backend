# GreenPOS Backend - SPEC

## Overview

Multi-tenant Restaurant POS System Backend in Go with GraphQL API.

## Architecture

```
greenpos-backend/
├── cmd/
│   └── server/           # Entry point
├── internal/
│   ├── config/          # Configuration
│   ├── database/        # Database connection
│   ├── graph/           # GraphQL resolvers
│   ├── middleware/      # Auth, tenant isolation
│   ├── model/           # Domain models
│   ├── repository/      # Data access layer
│   └── service/         # Business logic
├── pkg/
│   └── auth/            # JWT utilities
├── schema/              # GraphQL schema files
└── migrations/         # SQL migrations
```

## Multi-Tenancy Model

```
Tenant (Restaurant Chain/Owner)
  └── Branch (Individual restaurant location)
       ├── User (Employees)
       ├── Table (Restaurant tables)
       └── Order (Customer orders)
            └── OrderItem (Line items)
```

## Database Schema

### Tables

```sql
-- Tenants (restaurant owners/chains)
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    settings JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Branches (individual restaurant locations)
CREATE TABLE branches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    phone VARCHAR(50),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Users (employees within a branch)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('owner', 'admin', 'manager', 'waiter', 'kitchen')),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(branch_id, email)
);

-- Categories (menu categories)
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    sort_order INT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Products (menu items)
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    image_url TEXT,
    is_available BOOLEAN DEFAULT true,
    is_featured BOOLEAN DEFAULT false,
    preparation_time INT DEFAULT 15, -- minutes
    allergens TEXT[] DEFAULT '{}',
    rating DECIMAL(3,2) DEFAULT 0,
    review_count INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Tables (restaurant tables)
CREATE TABLE tables (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    number VARCHAR(20) NOT NULL,
    capacity INT DEFAULT 4,
    status VARCHAR(50) DEFAULT 'available' CHECK (status IN ('available', 'occupied', 'reserved')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(branch_id, number)
);

-- Orders
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    table_id UUID REFERENCES tables(id) ON DELETE SET NULL,
    user_id UUID NOT NULL REFERENCES users(id),
    customer_name VARCHAR(255),
    customer_phone VARCHAR(50),
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'preparing', 'ready', 'delivered', 'cancelled', 'paid')),
    subtotal DECIMAL(10,2) DEFAULT 0,
    tax DECIMAL(10,2) DEFAULT 0,
    discount DECIMAL(10,2) DEFAULT 0,
    total DECIMAL(10,2) DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Order Items
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity INT NOT NULL DEFAULT 1,
    unit_price DECIMAL(10,2) NOT NULL,
    total_price DECIMAL(10,2) NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Reservations
CREATE TABLE reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    table_id UUID REFERENCES tables(id) ON DELETE SET NULL,
    customer_name VARCHAR(255) NOT NULL,
    customer_phone VARCHAR(50),
    customer_email VARCHAR(255),
    guest_count INT DEFAULT 1,
    reservation_date DATE NOT NULL,
    reservation_time TIME NOT NULL,
    status VARCHAR(50) DEFAULT 'confirmed' CHECK (status IN ('pending', 'confirmed', 'cancelled', 'completed')),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_branches_tenant ON branches(tenant_id);
CREATE INDEX idx_users_branch ON users(branch_id);
CREATE INDEX idx_categories_branch ON categories(branch_id);
CREATE INDEX idx_products_branch ON products(branch_id);
CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_tables_branch ON tables(branch_id);
CREATE INDEX idx_orders_branch ON orders(branch_id);
CREATE INDEX idx_orders_table ON orders(table_id);
CREATE INDEX idx_orders_user ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_order_items_order ON order_items(order_id);
CREATE INDEX idx_reservations_branch ON reservations(branch_id);
CREATE INDEX idx_reservations_date ON reservations(reservation_date);
```

## GraphQL Schema

### Types

```graphql
scalar UUID
scalar Time
scalar Date
scalar JSON

type Tenant {
  id: UUID!
  name: String!
  slug: String!
  settings: JSON
  isActive: Boolean!
  branches: [Branch!]!
  createdAt: Time!
  updatedAt: Time!
}

type Branch {
  id: UUID!
  tenant: Tenant!
  name: String!
  address: String
  phone: String
  isActive: Boolean!
  users: [User!]!
  categories: [Category!]!
  products: [Product!]!
  tables: [Table!]!
  orders: [Order!]!
  reservations: [Reservation!]!
  createdAt: Time!
  updatedAt: Time!
}

type User {
  id: UUID!
  branch: Branch!
  email: String!
  name: String!
  role: UserRole!
  isActive: Boolean!
  createdAt: Time!
  updatedAt: Time!
}

enum UserRole {
  OWNER
  ADMIN
  MANAGER
  WAITER
  KITCHEN
}

type Category {
  id: UUID!
  branch: Branch!
  name: String!
  description: String
  sortOrder: Int!
  isActive: Boolean!
  products: [Product!]!
  createdAt: Time!
  updatedAt: Time!
}

type Product {
  id: UUID!
  branch: Branch!
  category: Category
  name: String!
  description: String
  price: Float!
  imageUrl: String
  isAvailable: Boolean!
  isFeatured: Boolean!
  preparationTime: Int!
  allergens: [String!]!
  rating: Float!
  reviewCount: Int!
  createdAt: Time!
  updatedAt: Time!
}

type Table {
  id: UUID!
  branch: Branch!
  number: String!
  capacity: Int!
  status: TableStatus!
  currentOrder: Order
  createdAt: Time!
  updatedAt: Time!
}

enum TableStatus {
  AVAILABLE
  OCCUPIED
  RESERVED
}

type Order {
  id: UUID!
  branch: Branch!
  table: Table
  user: User!
  customerName: String
  customerPhone: String
  status: OrderStatus!
  subtotal: Float!
  tax: Float!
  discount: Float!
  total: Float!
  notes: String
  items: [OrderItem!]!
  createdAt: Time!
  updatedAt: Time!
}

enum OrderStatus {
  PENDING
  ACCEPTED
  PREPARING
  READY
  DELIVERED
  CANCELLED
  PAID
}

type OrderItem {
  id: UUID!
  order: Order!
  product: Product!
  quantity: Int!
  unitPrice: Float!
  totalPrice: Float!
  notes: String
  createdAt: Time!
}

type Reservation {
  id: UUID!
  branch: Branch!
  table: Table
  customerName: String!
  customerPhone: String
  customerEmail: String
  guestCount: Int!
  reservationDate: Date!
  reservationTime: Time!
  status: ReservationStatus!
  notes: String
  createdAt: Time!
  updatedAt: Time!
}

enum ReservationStatus {
  PENDING
  CONFIRMED
  CANCELLED
  COMPLETED
}

# Inputs
input CreateTenantInput {
  name: String!
  slug: String!
  settings: JSON
}

input CreateBranchInput {
  tenantId: UUID!
  name: String!
  address: String
  phone: String
}

input CreateUserInput {
  branchId: UUID!
  email: String!
  password: String!
  name: String!
  role: UserRole!
}

input CreateCategoryInput {
  branchId: UUID!
  name: String!
  description: String
  sortOrder: Int
}

input CreateProductInput {
  branchId: UUID!
  categoryId: UUID
  name: String!
  description: String
  price: Float!
  imageUrl: String
  isAvailable: Boolean
  isFeatured: Boolean
  preparationTime: Int
  allergens: [String!]
}

input CreateTableInput {
  branchId: UUID!
  number: String!
  capacity: Int
}

input CreateOrderInput {
  branchId: UUID!
  tableId: UUID
  userId: UUID!
  customerName: String
  customerPhone: String
  notes: String
  items: [CreateOrderItemInput!]!
}

input CreateOrderItemInput {
  productId: UUID!
  quantity: Int!
  notes: String
}

input CreateReservationInput {
  branchId: UUID!
  tableId: UUID
  customerName: String!
  customerPhone: String
  customerEmail: String
  guestCount: Int!
  reservationDate: Date!
  reservationTime: Time!
  notes: String
}

# Auth
type AuthPayload {
  token: String!
  user: User!
}

# Queries
type Query {
  # Tenant queries
  me: User!
  tenant: Tenant!
  
  # Branch queries
  branches: [Branch!]!
  branch(id: UUID!): Branch
  
  # User queries
  users(branchId: UUID!): [User!]!
  user(id: UUID!): User
  
  # Category queries
  categories(branchId: UUID!): [Category!]!
  category(id: UUID!): Category
  
  # Product queries
  products(branchId: UUID!, categoryId: UUID): [Product!]!
  product(id: UUID!): Product
  featuredProducts(branchId: UUID!): [Product!]!
  
  # Table queries
  tables(branchId: UUID!): [Table!]!
  table(id: UUID!): Table
  
  # Order queries
  orders(branchId: UUID!, status: OrderStatus, limit: Int): [Order!]!
  order(id: UUID!): Order
  activeOrders(branchId: UUID!): [Order!]!
  
  # Reservation queries
  reservations(branchId: UUID!, date: Date): [Reservation!]!
  reservation(id: UUID!): Reservation
}

# Mutations
type Mutation {
  # Auth
  login(email: String!, password: String!, branchId: UUID!): AuthPayload!
  
  # Tenant
  createTenant(input: CreateTenantInput!): Tenant!
  updateTenant(id: UUID!, name: String, settings: JSON): Tenant!
  
  # Branch
  createBranch(input: CreateBranchInput!): Branch!
  updateBranch(id: UUID!, name: String, address: String, phone: String): Branch!
  
  # User
  createUser(input: CreateUserInput!): User!
  updateUser(id: UUID!, name: String, role: UserRole, isActive: Boolean): User!
  deleteUser(id: UUID!): Boolean!
  
  # Category
  createCategory(input: CreateCategoryInput!): Category!
  updateCategory(id: UUID!, name: String, description: String, sortOrder: Int, isActive: Boolean): Category!
  deleteCategory(id: UUID!): Boolean!
  reorderCategories(branchId: UUID!, categoryIds: [UUID!]!): [Category!]!
  
  # Product
  createProduct(input: CreateProductInput!): Product!
  updateProduct(id: UUID!, input: CreateProductInput!): Product!
  deleteProduct(id: UUID!): Boolean!
  toggleProductAvailability(id: UUID!): Product!
  
  # Table
  createTable(input: CreateTableInput!): Table!
  updateTable(id: UUID!, number: String, capacity: Int): Table!
  deleteTable(id: UUID!): Boolean!
  updateTableStatus(id: UUID!, status: TableStatus!): Table!
  
  # Order
  createOrder(input: CreateOrderInput!): Order!
  updateOrderStatus(id: UUID!, status: OrderStatus!): Order!
  cancelOrder(id: UUID!): Order!
  addOrderItem(orderId: UUID!, input: CreateOrderItemInput!): OrderItem!
  removeOrderItem(id: UUID!): Boolean!
  
  # Reservation
  createReservation(input: CreateReservationInput!): Reservation!
  updateReservation(id: UUID!, input: CreateReservationInput!): Reservation!
  cancelReservation(id: UUID!): Reservation!
}
```

## API Endpoints

### REST (for health, webhooks, etc.)
- `GET /health` - Health check
- `POST /webhooks/stripe` - Payment webhooks

### GraphQL
- `POST /graphql` - Main GraphQL endpoint
- `GET /graphql/playground` - GraphQL Playground (dev only)

## Environment Variables

```env
# Server
PORT=8080
ENV=development

# Database
DATABASE_URL=postgres://user:pass@localhost:5432/greenpos?sslmode=disable

# JWT
JWT_SECRET=your-secret-key-here
JWT_EXPIRY=24h

# CORS
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
```
