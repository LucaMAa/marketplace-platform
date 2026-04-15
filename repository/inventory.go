package repository

import (
	"errors"
	"marketplace-platform/models"

	"gorm.io/gorm"
)

type InventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

func (r *InventoryRepository) Create(inventory *models.Inventory) error {
	return r.db.Create(inventory).Error
}

func (r *InventoryRepository) GetByMerchantAndProduct(merchantID, productID uint) (*models.Inventory, error) {
	var inventory models.Inventory
	if err := r.db.
		Preload("Product").
		Where("merchant_id = ? AND product_id = ?", merchantID, productID).
		First(&inventory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &inventory, nil
}

func (r *InventoryRepository) GetByMerchantID(merchantID uint) ([]models.Inventory, error) {
	var inventories []models.Inventory
	err := r.db.
		Preload("Product").
		Where("merchant_id = ?", merchantID).
		Find(&inventories).Error
	return inventories, err
}

func (r *InventoryRepository) Update(inventory *models.Inventory) error {
	return r.db.Save(inventory).Error
}

func (r *InventoryRepository) GetByProductCode(productCode string) ([]models.Inventory, error) {
	var inventories []models.Inventory
	err := r.db.
		Joins("JOIN products ON inventories.product_id = products.id").
		Where("products.code = ?", productCode).
		Preload("Product").
		Preload("Merchant").
		Find(&inventories).Error
	return inventories, err
}
