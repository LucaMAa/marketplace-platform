package repository

import (
	"errors"
	"marketplace-platform/models"

	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

func (r *ProductRepository) GetByID(id uint) (*models.Product, error) {
	var product models.Product
	if err := r.db.First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) GetByCode(code string) (*models.Product, error) {
	var product models.Product
	if err := r.db.Where("code = ?", code).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) Search(query string) ([]models.Product, error) {
	var products []models.Product
	err := r.db.
		Where("name ILIKE ? OR code ILIKE ? OR description ILIKE ?", "%"+query+"%", "%"+query+"%", "%"+query+"%"). //ANCHE NO
		Find(&products).Error
	return products, err
}

func (r *ProductRepository) GetByCategory(categoryID uint) ([]models.Product, error) {
	var products []models.Product
	err := r.db.
		Where("category_id = ?", categoryID).
		Find(&products).Error
	return products, err
}
