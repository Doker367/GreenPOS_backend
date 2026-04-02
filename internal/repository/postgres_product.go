package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
	"gorm.io/gorm"
)

// GormProductRepository is a GORM-backed ProductRepository
type GormProductRepository struct {
	db *gorm.DB
}

type GormProductModel struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BranchID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"branchId"`
	CategoryID      *uuid.UUID `gorm:"type:uuid;index" json:"categoryId"`
	Name            string     `gorm:"type:varchar(255);not null" json:"name"`
	Description     string     `gorm:"type:text" json:"description"`
	Price           float64    `gorm:"type:decimal(10,2);not null" json:"price"`
	ImageURL        string     `gorm:"type:text" json:"imageUrl"`
	IsAvailable     bool       `gorm:"default:true" json:"isAvailable"`
	IsFeatured      bool       `gorm:"default:false" json:"isFeatured"`
	PreparationTime int        `gorm:"default:15" json:"preparationTime"`
	Allergens       []string   `gorm:"type:text[];default:'{}'" json:"allergens"`
	Rating          float64    `gorm:"type:decimal(3,2);default:0" json:"rating"`
	ReviewCount     int        `gorm:"default:0" json:"reviewCount"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (GormProductModel) TableName() string { return "products" }

func toProductModel(m *GormProductModel) *model.Product {
	return &model.Product{
		ID:              m.ID,
		BranchID:        m.BranchID,
		CategoryID:      m.CategoryID,
		Name:            m.Name,
		Description:     m.Description,
		Price:           m.Price,
		ImageURL:        m.ImageURL,
		IsAvailable:     m.IsAvailable,
		IsFeatured:      m.IsFeatured,
		PreparationTime: m.PreparationTime,
		Allergens:       m.Allergens,
		Rating:          m.Rating,
		ReviewCount:     m.ReviewCount,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

func toProductGorm(m *model.Product) *GormProductModel {
	return &GormProductModel{
		ID:              m.ID,
		BranchID:        m.BranchID,
		CategoryID:      m.CategoryID,
		Name:            m.Name,
		Description:     m.Description,
		Price:           m.Price,
		ImageURL:        m.ImageURL,
		IsAvailable:     m.IsAvailable,
		IsFeatured:      m.IsFeatured,
		PreparationTime: m.PreparationTime,
		Allergens:       m.Allergens,
		Rating:          m.Rating,
		ReviewCount:     m.ReviewCount,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

// NewGormProductRepository creates a GORM-backed product repository
func NewGormProductRepository(db *gorm.DB) *GormProductRepository {
	return &GormProductRepository{db: db}
}

func (r *GormProductRepository) Create(ctx context.Context, product *model.Product) error {
	if product.ID == uuid.Nil {
		product.ID = uuid.New()
	}
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(toProductGorm(product)).Error
}

func (r *GormProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	var m GormProductModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toProductModel(&m), nil
}

func (r *GormProductRepository) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Product, error) {
	var models []GormProductModel
	if err := r.db.WithContext(ctx).
		Where("branch_id = ?", branchID).
		Order("name ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]model.Product, len(models))
	for i := range models {
		result[i] = *toProductModel(&models[i])
	}
	return result, nil
}

func (r *GormProductRepository) GetByCategory(ctx context.Context, branchID, categoryID uuid.UUID) ([]model.Product, error) {
	var models []GormProductModel
	if err := r.db.WithContext(ctx).
		Where("branch_id = ? AND category_id = ?", branchID, categoryID).
		Order("name ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]model.Product, len(models))
	for i := range models {
		result[i] = *toProductModel(&models[i])
	}
	return result, nil
}

func (r *GormProductRepository) GetAvailable(ctx context.Context, branchID uuid.UUID) ([]model.Product, error) {
	var models []GormProductModel
	if err := r.db.WithContext(ctx).
		Where("branch_id = ? AND is_available = ?", branchID, true).
		Order("name ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]model.Product, len(models))
	for i := range models {
		result[i] = *toProductModel(&models[i])
	}
	return result, nil
}

func (r *GormProductRepository) GetFeatured(ctx context.Context, branchID uuid.UUID) ([]model.Product, error) {
	var models []GormProductModel
	if err := r.db.WithContext(ctx).
		Where("branch_id = ? AND is_featured = ? AND is_available = ?", branchID, true, true).
		Order("rating DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]model.Product, len(models))
	for i := range models {
		result[i] = *toProductModel(&models[i])
	}
	return result, nil
}

func (r *GormProductRepository) Update(ctx context.Context, product *model.Product) error {
	product.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(toProductGorm(product)).Error
}

func (r *GormProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&GormProductModel{}).
		Where("id = ?", id).
		Update("is_available", false).Error
}

func (r *GormProductRepository) ToggleAvailability(ctx context.Context, id uuid.UUID, isAvailable bool) error {
	return r.db.WithContext(ctx).
		Model(&GormProductModel{}).
		Where("id = ?", id).
		Update("is_available", isAvailable).Error
}

func (r *GormProductRepository) Seed(ctx context.Context) error {
	var count int64
	r.db.WithContext(ctx).Model(&GormProductModel{}).Count(&count)
	if count > 0 {
		return nil
	}

	cat1 := uuid.MustParse("44444444-4444-4444-4444-444444444441")
	cat2 := uuid.MustParse("44444444-4444-4444-4444-444444444442")
	cat3 := uuid.MustParse("44444444-4444-4444-4444-444444444443")
	cat4 := uuid.MustParse("44444444-4444-4444-4444-444444444444")

	products := []*model.Product{
		{
			ID:              uuid.MustParse("55555555-5555-5555-5555-555555555551"),
			BranchID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			CategoryID:      &cat1,
			Name:            "Bruschetta",
			Description:     "Toasted bread with fresh tomatoes and basil",
			Price:           8.99,
			IsAvailable:     true,
			IsFeatured:      true,
			PreparationTime: 10,
			Allergens:       []string{"gluten"},
			Rating:          4.5,
			ReviewCount:     120,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              uuid.MustParse("55555555-5555-5555-5555-555555555552"),
			BranchID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			CategoryID:      &cat1,
			Name:            "Garlic Bread",
			Description:     "Crusty bread with garlic butter",
			Price:           5.99,
			IsAvailable:     true,
			PreparationTime: 5,
			Allergens:       []string{"gluten", "dairy"},
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              uuid.MustParse("55555555-5555-5555-5555-555555555553"),
			BranchID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			CategoryID:      &cat2,
			Name:            "Grilled Salmon",
			Description:     "Fresh Atlantic salmon with herbs",
			Price:           24.99,
			IsAvailable:     true,
			IsFeatured:      true,
			PreparationTime: 20,
			Allergens:       []string{"fish"},
			Rating:          4.8,
			ReviewCount:     95,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              uuid.MustParse("55555555-5555-5555-5555-555555555554"),
			BranchID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			CategoryID:      &cat2,
			Name:            "Ribeye Steak",
			Description:     "Prime cut with garlic butter",
			Price:           34.99,
			IsAvailable:     true,
			PreparationTime: 25,
			Allergens:       []string{"dairy"},
			Rating:          4.9,
			ReviewCount:     150,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              uuid.MustParse("55555555-5555-5555-5555-555555555555"),
			BranchID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			CategoryID:      &cat3,
			Name:            "Chocolate Cake",
			Description:     "Rich dark chocolate layer cake",
			Price:           9.99,
			IsAvailable:     true,
			PreparationTime: 5,
			Allergens:       []string{"gluten", "dairy", "eggs"},
			Rating:          4.7,
			ReviewCount:     80,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              uuid.MustParse("55555555-5555-5555-5555-555555555556"),
			BranchID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			CategoryID:      &cat4,
			Name:            "Fresh Lemonade",
			Description:     "Homemade with fresh lemons",
			Price:           4.99,
			IsAvailable:     true,
			PreparationTime: 3,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	for _, p := range products {
		if err := r.Create(ctx, p); err != nil {
			return err
		}
	}
	return nil
}
