package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/greenpos/backend/internal/model"
)

type Repositories struct {
	Tenants      *TenantRepository
	Branches     *BranchRepository
	Users        *UserRepository
	Categories   *CategoryRepository
	Products     *ProductRepository
	Tables       *TableRepository
	Orders       *OrderRepository
	Reservations *ReservationRepository
}

func NewRepositories(db *pgxpool.Pool) *Repositories {
	return &Repositories{
		Tenants:      NewTenantRepository(db),
		Branches:     NewBranchRepository(db),
		Users:        NewUserRepository(db),
		Categories:   NewCategoryRepository(db),
		Products:     NewProductRepository(db),
		Tables:       NewTableRepository(db),
		Orders:       NewOrderRepository(db),
		Reservations: NewReservationRepository(db),
	}
}

// ============ TENANT REPOSITORY ============

type TenantRepository struct {
	db *pgxpool.Pool
}

func NewTenantRepository(db *pgxpool.Pool) *TenantRepository {
	return &TenantRepository{db: db}
}

func (r *TenantRepository) Create(ctx context.Context, tenant *model.Tenant) error {
	query := `
		INSERT INTO tenants (id, name, slug, settings, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, tenant.ID, tenant.Name, tenant.Slug, tenant.Settings, tenant.IsActive)
	return err
}

func (r *TenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Tenant, error) {
	query := `SELECT id, name, slug, settings, is_active, created_at, updated_at FROM tenants WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	var t model.Tenant
	err := row.Scan(&t.ID, &t.Name, &t.Slug, &t.Settings, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TenantRepository) GetBySlug(ctx context.Context, slug string) (*model.Tenant, error) {
	query := `SELECT id, name, slug, settings, is_active, created_at, updated_at FROM tenants WHERE slug = $1`
	row := r.db.QueryRow(ctx, query, slug)
	var t model.Tenant
	err := row.Scan(&t.ID, &t.Name, &t.Slug, &t.Settings, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TenantRepository) Update(ctx context.Context, tenant *model.Tenant) error {
	query := `UPDATE tenants SET name = $2, slug = $3, settings = $4, is_active = $5, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, tenant.ID, tenant.Name, tenant.Slug, tenant.Settings, tenant.IsActive)
	return err
}

func (r *TenantRepository) List(ctx context.Context) ([]model.Tenant, error) {
	query := `SELECT id, name, slug, settings, is_active, created_at, updated_at FROM tenants ORDER BY name`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []model.Tenant
	for rows.Next() {
		var t model.Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Settings, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tenants = append(tenants, t)
	}
	return tenants, nil
}

// ============ BRANCH REPOSITORY ============

type BranchRepository struct {
	db *pgxpool.Pool
}

func NewBranchRepository(db *pgxpool.Pool) *BranchRepository {
	return &BranchRepository{db: db}
}

func (r *BranchRepository) Create(ctx context.Context, branch *model.Branch) error {
	query := `
		INSERT INTO branches (id, tenant_id, name, address, phone, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, branch.ID, branch.TenantID, branch.Name, branch.Address, branch.Phone, branch.IsActive)
	return err
}

func (r *BranchRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Branch, error) {
	query := `SELECT id, tenant_id, name, address, phone, is_active, created_at, updated_at FROM branches WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	var b model.Branch
	err := row.Scan(&b.ID, &b.TenantID, &b.Name, &b.Address, &b.Phone, &b.IsActive, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BranchRepository) GetByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.Branch, error) {
	query := `SELECT id, tenant_id, name, address, phone, is_active, created_at, updated_at 
	          FROM branches WHERE tenant_id = $1 ORDER BY name`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var branches []model.Branch
	for rows.Next() {
		var b model.Branch
		if err := rows.Scan(&b.ID, &b.TenantID, &b.Name, &b.Address, &b.Phone, &b.IsActive, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		branches = append(branches, b)
	}
	return branches, nil
}

// ListByTenant is an alias for GetByTenant
func (r *BranchRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.Branch, error) {
	return r.GetByTenant(ctx, tenantID)
}

func (r *BranchRepository) Update(ctx context.Context, branch *model.Branch) error {
	query := `UPDATE branches SET name = $2, address = $3, phone = $4, is_active = $5, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, branch.ID, branch.Name, branch.Address, branch.Phone, branch.IsActive)
	return err
}

// ============ USER REPOSITORY ============

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, branch_id, email, password_hash, name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, user.ID, user.BranchID, user.Email, user.PasswordHash, user.Name, user.Role, user.IsActive)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `SELECT id, branch_id, email, password_hash, name, role, is_active, created_at, updated_at FROM users WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	var u model.User
	err := row.Scan(&u.ID, &u.BranchID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.User, error) {
	query := `SELECT id, branch_id, email, password_hash, name, role, is_active, created_at, updated_at 
	          FROM users WHERE branch_id = $1 AND is_active = true`
	rows, err := r.db.Query(ctx, query, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.BranchID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// ListByBranch is an alias for GetByBranch
func (r *UserRepository) ListByBranch(ctx context.Context, branchID uuid.UUID) ([]model.User, error) {
	return r.GetByBranch(ctx, branchID)
}

func (r *UserRepository) GetByEmail(ctx context.Context, branchID uuid.UUID, email string) (*model.User, error) {
	query := `SELECT id, branch_id, email, password_hash, name, role, is_active, created_at, updated_at 
	          FROM users WHERE branch_id = $1 AND email = $2`
	var u model.User
	err := r.db.QueryRow(ctx, query, branchID, email).Scan(
		&u.ID, &u.BranchID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	query := `UPDATE users SET email = $2, name = $3, role = $4, is_active = $5, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, user.ID, user.Email, user.Name, user.Role, user.IsActive)
	return err
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, passwordHash)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// ============ CATEGORY REPOSITORY ============

type CategoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, category *model.Category) error {
	query := `
		INSERT INTO categories (id, branch_id, name, description, sort_order, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, category.ID, category.BranchID, category.Name, category.Description, category.SortOrder, category.IsActive)
	return err
}

func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	query := `SELECT id, branch_id, name, description, sort_order, is_active, created_at, updated_at FROM categories WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	var c model.Category
	err := row.Scan(&c.ID, &c.BranchID, &c.Name, &c.Description, &c.SortOrder, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CategoryRepository) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Category, error) {
	query := `SELECT id, branch_id, name, description, sort_order, is_active, created_at, updated_at 
	          FROM categories WHERE branch_id = $1 ORDER BY sort_order, name`
	rows, err := r.db.Query(ctx, query, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.BranchID, &c.Name, &c.Description, &c.SortOrder, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

// ListByBranch is an alias for GetByBranch
func (r *CategoryRepository) ListByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Category, error) {
	return r.GetByBranch(ctx, branchID)
}

func (r *CategoryRepository) Update(ctx context.Context, category *model.Category) error {
	query := `UPDATE categories SET name = $2, description = $3, sort_order = $4, is_active = $5, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, category.ID, category.Name, category.Description, category.SortOrder, category.IsActive)
	return err
}

func (r *CategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE categories SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *CategoryRepository) Reorder(ctx context.Context, branchID uuid.UUID, categoryOrders []struct {
	ID        uuid.UUID
	SortOrder int
}) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, co := range categoryOrders {
		_, err := tx.Exec(ctx, `UPDATE categories SET sort_order = $2, updated_at = NOW() WHERE id = $1 AND branch_id = $3`, co.ID, co.SortOrder, branchID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *CategoryRepository) GetMaxSortOrder(ctx context.Context, branchID uuid.UUID) (int, error) {
	query := `SELECT COALESCE(MAX(sort_order), 0) FROM categories WHERE branch_id = $1`
	var maxOrder int
	err := r.db.QueryRow(ctx, query, branchID).Scan(&maxOrder)
	return maxOrder, err
}

// ============ PRODUCT REPOSITORY ============

type ProductRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, product *model.Product) error {
	query := `
		INSERT INTO products (id, branch_id, category_id, name, description, price, image_url, 
		                      is_available, is_featured, preparation_time, allergens, rating, review_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, product.ID, product.BranchID, product.CategoryID, product.Name,
		product.Description, product.Price, product.ImageURL, product.IsAvailable, product.IsFeatured,
		product.PreparationTime, product.Allergens, product.Rating, product.ReviewCount)
	return err
}

func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	query := `SELECT id, branch_id, category_id, name, description, price, image_url, is_available, 
	                 is_featured, preparation_time, allergens, rating, review_count, created_at, updated_at 
	          FROM products WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	var p model.Product
	err := row.Scan(&p.ID, &p.BranchID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &p.ImageURL,
		&p.IsAvailable, &p.IsFeatured, &p.PreparationTime, &p.Allergens, &p.Rating, &p.ReviewCount, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepository) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Product, error) {
	query := `SELECT id, branch_id, category_id, name, description, price, image_url, is_available, 
	                 is_featured, preparation_time, allergens, rating, review_count, created_at, updated_at 
	          FROM products WHERE branch_id = $1 ORDER BY name`
	rows, err := r.db.Query(ctx, query, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.BranchID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &p.ImageURL,
			&p.IsAvailable, &p.IsFeatured, &p.PreparationTime, &p.Allergens, &p.Rating, &p.ReviewCount, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

// ListByBranch is an alias for GetByBranch
func (r *ProductRepository) ListByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Product, error) {
	return r.GetByBranch(ctx, branchID)
}

func (r *ProductRepository) GetByCategory(ctx context.Context, branchID, categoryID uuid.UUID) ([]model.Product, error) {
	query := `SELECT id, branch_id, category_id, name, description, price, image_url, is_available, 
	                 is_featured, preparation_time, allergens, rating, review_count, created_at, updated_at 
	          FROM products WHERE branch_id = $1 AND category_id = $2 ORDER BY name`
	rows, err := r.db.Query(ctx, query, branchID, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.BranchID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &p.ImageURL,
			&p.IsAvailable, &p.IsFeatured, &p.PreparationTime, &p.Allergens, &p.Rating, &p.ReviewCount, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

// ListByCategory is an alias for GetByCategory
func (r *ProductRepository) ListByCategory(ctx context.Context, branchID, categoryID uuid.UUID) ([]model.Product, error) {
	return r.GetByCategory(ctx, branchID, categoryID)
}

func (r *ProductRepository) GetFeatured(ctx context.Context, branchID uuid.UUID) ([]model.Product, error) {
	query := `SELECT id, branch_id, category_id, name, description, price, image_url, is_available, 
	                 is_featured, preparation_time, allergens, rating, review_count, created_at, updated_at 
	          FROM products WHERE branch_id = $1 AND is_featured = true AND is_available = true ORDER BY rating DESC, name`
	rows, err := r.db.Query(ctx, query, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.BranchID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &p.ImageURL,
			&p.IsAvailable, &p.IsFeatured, &p.PreparationTime, &p.Allergens, &p.Rating, &p.ReviewCount, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *ProductRepository) GetAvailable(ctx context.Context, branchID uuid.UUID) ([]model.Product, error) {
	query := `SELECT id, branch_id, category_id, name, description, price, image_url, is_available, 
	                 is_featured, preparation_time, allergens, rating, review_count, created_at, updated_at 
	          FROM products WHERE branch_id = $1 AND is_available = true ORDER BY name`
	rows, err := r.db.Query(ctx, query, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.BranchID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &p.ImageURL,
			&p.IsAvailable, &p.IsFeatured, &p.PreparationTime, &p.Allergens, &p.Rating, &p.ReviewCount, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *ProductRepository) Update(ctx context.Context, product *model.Product) error {
	query := `UPDATE products SET category_id = $2, name = $3, description = $4, price = $5, 
	          image_url = $6, is_available = $7, is_featured = $8, preparation_time = $9, allergens = $10, 
	          updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, product.ID, product.CategoryID, product.Name, product.Description,
		product.Price, product.ImageURL, product.IsAvailable, product.IsFeatured, product.PreparationTime, product.Allergens)
	return err
}

func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE products SET is_available = false, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *ProductRepository) ToggleAvailability(ctx context.Context, id uuid.UUID, isAvailable bool) error {
	query := `UPDATE products SET is_available = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, isAvailable)
	return err
}

// ============ TABLE REPOSITORY ============

type TableRepository struct {
	db *pgxpool.Pool
}

func NewTableRepository(db *pgxpool.Pool) *TableRepository {
	return &TableRepository{db: db}
}

func (r *TableRepository) Create(ctx context.Context, table *model.RestaurantTable) error {
	query := `
		INSERT INTO tables (id, branch_id, number, capacity, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, table.ID, table.BranchID, table.Number, table.Capacity, table.Status)
	return err
}

func (r *TableRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.RestaurantTable, error) {
	query := `SELECT id, branch_id, number, capacity, status, created_at, updated_at FROM tables WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	var t model.RestaurantTable
	err := row.Scan(&t.ID, &t.BranchID, &t.Number, &t.Capacity, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TableRepository) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.RestaurantTable, error) {
	query := `SELECT id, branch_id, number, capacity, status, created_at, updated_at 
	          FROM tables WHERE branch_id = $1 ORDER BY number`
	rows, err := r.db.Query(ctx, query, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []model.RestaurantTable
	for rows.Next() {
		var t model.RestaurantTable
		if err := rows.Scan(&t.ID, &t.BranchID, &t.Number, &t.Capacity, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}
	return tables, nil
}

// ListByBranch is an alias for GetByBranch
func (r *TableRepository) ListByBranch(ctx context.Context, branchID uuid.UUID) ([]model.RestaurantTable, error) {
	return r.GetByBranch(ctx, branchID)
}

func (r *TableRepository) Update(ctx context.Context, table *model.RestaurantTable) error {
	query := `UPDATE tables SET number = $2, capacity = $3, status = $4, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, table.ID, table.Number, table.Capacity, table.Status)
	return err
}

func (r *TableRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.TableStatus) error {
	query := `UPDATE tables SET status = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, status)
	return err
}

func (r *TableRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE tables SET status = 'UNAVAILABLE', updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *TableRepository) GetAvailable(ctx context.Context, branchID uuid.UUID) ([]model.RestaurantTable, error) {
	query := `SELECT id, branch_id, number, capacity, status, created_at, updated_at 
	          FROM tables WHERE branch_id = $1 AND status = 'AVAILABLE' ORDER BY number`
	rows, err := r.db.Query(ctx, query, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []model.RestaurantTable
	for rows.Next() {
		var t model.RestaurantTable
		if err := rows.Scan(&t.ID, &t.BranchID, &t.Number, &t.Capacity, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}
	return tables, nil
}

// ============ ORDER REPOSITORY ============

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order *model.Order) error {
	query := `
		INSERT INTO orders (id, branch_id, table_id, user_id, customer_name, customer_phone, status, 
		                    subtotal, tax, discount, total, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, order.ID, order.BranchID, order.TableID, order.UserID, order.CustomerName,
		order.CustomerPhone, order.Status, order.Subtotal, order.Tax, order.Discount, order.Total, order.Notes)
	return err
}

func (r *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	query := `SELECT id, branch_id, table_id, user_id, customer_name, customer_phone, status, 
	                 subtotal, tax, discount, total, notes, created_at, updated_at 
	          FROM orders WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	var o model.Order
	err := row.Scan(&o.ID, &o.BranchID, &o.TableID, &o.UserID, &o.CustomerName, &o.CustomerPhone,
		&o.Status, &o.Subtotal, &o.Tax, &o.Discount, &o.Total, &o.Notes, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepository) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Order, error) {
	query := `SELECT id, branch_id, table_id, user_id, customer_name, customer_phone, status, 
	                 subtotal, tax, discount, total, notes, created_at, updated_at 
	          FROM orders WHERE branch_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.BranchID, &o.TableID, &o.UserID, &o.CustomerName, &o.CustomerPhone,
			&o.Status, &o.Subtotal, &o.Tax, &o.Discount, &o.Total, &o.Notes, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

// ListByBranch returns orders for a branch with optional status filter and limit
func (r *OrderRepository) ListByBranch(ctx context.Context, branchID uuid.UUID, status *model.OrderStatus, limit *int) ([]model.Order, error) {
	var query string
	var args []interface{}
	
	if status != nil {
		query = `SELECT id, branch_id, table_id, user_id, customer_name, customer_phone, status, 
		                 subtotal, tax, discount, total, notes, created_at, updated_at 
		          FROM orders WHERE branch_id = $1 AND status = $2 ORDER BY created_at DESC`
		if limit != nil {
			query += fmt.Sprintf(` LIMIT %d`, *limit)
		}
		args = []interface{}{branchID, *status}
	} else {
		query = `SELECT id, branch_id, table_id, user_id, customer_name, customer_phone, status, 
		                 subtotal, tax, discount, total, notes, created_at, updated_at 
		          FROM orders WHERE branch_id = $1 ORDER BY created_at DESC`
		if limit != nil {
			query += fmt.Sprintf(` LIMIT %d`, *limit)
		}
		args = []interface{}{branchID}
	}
	
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.BranchID, &o.TableID, &o.UserID, &o.CustomerName, &o.CustomerPhone,
			&o.Status, &o.Subtotal, &o.Tax, &o.Discount, &o.Total, &o.Notes, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (r *OrderRepository) Update(ctx context.Context, order *model.Order) error {
	query := `UPDATE orders SET table_id = $2, customer_name = $3, customer_phone = $4, status = $5,
	          subtotal = $6, tax = $7, discount = $8, total = $9, notes = $10, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, order.ID, order.TableID, order.CustomerName, order.CustomerPhone,
		order.Status, order.Subtotal, order.Tax, order.Discount, order.Total, order.Notes)
	return err
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus) error {
	query := `UPDATE orders SET status = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, status)
	return err
}

func (r *OrderRepository) AddItem(ctx context.Context, item *model.OrderItem) error {
	query := `
		INSERT INTO order_items (id, order_id, product_id, quantity, unit_price, total_price, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`
	_, err := r.db.Exec(ctx, query, item.ID, item.OrderID, item.ProductID, item.Quantity, item.UnitPrice, item.TotalPrice, item.Notes)
	return err
}

func (r *OrderRepository) GetItems(ctx context.Context, orderID uuid.UUID) ([]model.OrderItem, error) {
	query := `SELECT id, order_id, product_id, quantity, unit_price, total_price, notes, created_at 
	          FROM order_items WHERE order_id = $1`
	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.OrderItem
	for rows.Next() {
		var i model.OrderItem
		if err := rows.Scan(&i.ID, &i.OrderID, &i.ProductID, &i.Quantity, &i.UnitPrice, &i.TotalPrice, &i.Notes, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}

func (r *OrderRepository) GetItem(ctx context.Context, itemID uuid.UUID) (*model.OrderItem, error) {
	query := `SELECT id, order_id, product_id, quantity, unit_price, total_price, notes, created_at 
	          FROM order_items WHERE id = $1`
	row := r.db.QueryRow(ctx, query, itemID)
	var i model.OrderItem
	err := row.Scan(&i.ID, &i.OrderID, &i.ProductID, &i.Quantity, &i.UnitPrice, &i.TotalPrice, &i.Notes, &i.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *OrderRepository) RemoveItem(ctx context.Context, itemID uuid.UUID) error {
	query := `DELETE FROM order_items WHERE id = $1`
	_, err := r.db.Exec(ctx, query, itemID)
	return err
}

func (r *OrderRepository) GetActiveByTable(ctx context.Context, tableID uuid.UUID) (*model.Order, error) {
	query := `SELECT id, branch_id, table_id, user_id, customer_name, customer_phone, status, 
	                 subtotal, tax, discount, total, notes, created_at, updated_at 
	          FROM orders WHERE table_id = $1 AND status NOT IN ('PAID', 'CANCELLED') LIMIT 1`
	row := r.db.QueryRow(ctx, query, tableID)
	var o model.Order
	err := row.Scan(&o.ID, &o.BranchID, &o.TableID, &o.UserID, &o.CustomerName, &o.CustomerPhone,
		&o.Status, &o.Subtotal, &o.Tax, &o.Discount, &o.Total, &o.Notes, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// GetActive returns all active orders for a branch (not cancelled/paid)
func (r *OrderRepository) GetActive(ctx context.Context, branchID uuid.UUID) ([]*model.Order, error) {
	query := `SELECT id, branch_id, table_id, user_id, customer_name, customer_phone, status, 
	                 subtotal, tax, discount, total, notes, created_at, updated_at 
	          FROM orders WHERE branch_id = $1 AND status NOT IN ('PAID', 'CANCELLED') ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.BranchID, &o.TableID, &o.UserID, &o.CustomerName, &o.CustomerPhone,
			&o.Status, &o.Subtotal, &o.Tax, &o.Discount, &o.Total, &o.Notes, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}
	return orders, nil
}

// ============ RESERVATION REPOSITORY ============

type ReservationRepository struct {
	db *pgxpool.Pool
}

func NewReservationRepository(db *pgxpool.Pool) *ReservationRepository {
	return &ReservationRepository{db: db}
}

func (r *ReservationRepository) Create(ctx context.Context, reservation *model.Reservation) error {
	query := `
		INSERT INTO reservations (id, branch_id, table_id, customer_name, customer_phone, customer_email, 
		                          guest_count, reservation_date, reservation_time, status, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, reservation.ID, reservation.BranchID, reservation.TableID,
		reservation.CustomerName, reservation.CustomerPhone, reservation.CustomerEmail,
		reservation.GuestCount, reservation.ReservationDate, reservation.ReservationTime,
		reservation.Status, reservation.Notes)
	return err
}

func (r *ReservationRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Reservation, error) {
	query := `SELECT id, branch_id, table_id, customer_name, customer_phone, customer_email, 
	                 guest_count, reservation_date, reservation_time, status, notes, created_at, updated_at 
	          FROM reservations WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	var res model.Reservation
	err := row.Scan(&res.ID, &res.BranchID, &res.TableID, &res.CustomerName, &res.CustomerPhone,
		&res.CustomerEmail, &res.GuestCount, &res.ReservationDate, &res.ReservationTime,
		&res.Status, &res.Notes, &res.CreatedAt, &res.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *ReservationRepository) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Reservation, error) {
	query := `SELECT id, branch_id, table_id, customer_name, customer_phone, customer_email, 
	                 guest_count, reservation_date, reservation_time, status, notes, created_at, updated_at 
	          FROM reservations WHERE branch_id = $1 ORDER BY reservation_date, reservation_time`
	rows, err := r.db.Query(ctx, query, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reservations []model.Reservation
	for rows.Next() {
		var res model.Reservation
		if err := rows.Scan(&res.ID, &res.BranchID, &res.TableID, &res.CustomerName, &res.CustomerPhone,
			&res.CustomerEmail, &res.GuestCount, &res.ReservationDate, &res.ReservationTime,
			&res.Status, &res.Notes, &res.CreatedAt, &res.UpdatedAt); err != nil {
			return nil, err
		}
		reservations = append(reservations, res)
	}
	return reservations, nil
}

// ListByBranch is an alias for GetByBranch
// ListByBranch returns reservations for a branch with optional date filter
func (r *ReservationRepository) ListByBranch(ctx context.Context, branchID uuid.UUID, date *time.Time) ([]model.Reservation, error) {
	if date != nil {
		return r.GetByDate(ctx, branchID, date.Format("2006-01-02"))
	}
	return r.GetByBranch(ctx, branchID)
}

func (r *ReservationRepository) GetByDate(ctx context.Context, branchID uuid.UUID, date string) ([]model.Reservation, error) {
	query := `SELECT id, branch_id, table_id, customer_name, customer_phone, customer_email, 
	                 guest_count, reservation_date, reservation_time, status, notes, created_at, updated_at 
	          FROM reservations WHERE branch_id = $1 AND reservation_date = $2 
	          ORDER BY reservation_time`
	rows, err := r.db.Query(ctx, query, branchID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reservations []model.Reservation
	for rows.Next() {
		var res model.Reservation
		if err := rows.Scan(&res.ID, &res.BranchID, &res.TableID, &res.CustomerName, &res.CustomerPhone,
			&res.CustomerEmail, &res.GuestCount, &res.ReservationDate, &res.ReservationTime,
			&res.Status, &res.Notes, &res.CreatedAt, &res.UpdatedAt); err != nil {
			return nil, err
		}
		reservations = append(reservations, res)
	}
	return reservations, nil
}

func (r *ReservationRepository) Update(ctx context.Context, reservation *model.Reservation) error {
	query := `UPDATE reservations SET table_id = $2, customer_name = $3, customer_phone = $4, 
	          customer_email = $5, guest_count = $6, reservation_date = $7, reservation_time = $8, 
	          status = $9, notes = $10, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, reservation.ID, reservation.TableID, reservation.CustomerName,
		reservation.CustomerPhone, reservation.CustomerEmail, reservation.GuestCount,
		reservation.ReservationDate, reservation.ReservationTime, reservation.Status, reservation.Notes)
	return err
}

func (r *ReservationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.ReservationStatus) error {
	query := `UPDATE reservations SET status = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, status)
	return err
}
