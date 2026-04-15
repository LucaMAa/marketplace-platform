package repository

import (
	"errors"
	"marketplace-platform/models"
	"time"

	"gorm.io/gorm"
)

type MerchantRepository struct {
	db *gorm.DB
}

func NewMerchantRepository(db *gorm.DB) *MerchantRepository {
	return &MerchantRepository{db: db}
}

func (r *MerchantRepository) Create(merchant *models.Merchant) error {
	return r.db.Create(merchant).Error
}

func (r *MerchantRepository) GetByEmail(email string) (*models.Merchant, error) {
	var merchant models.Merchant
	if err := r.db.Where("email = ?", email).First(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &merchant, nil
}

func (r *MerchantRepository) GetByID(id uint) (*models.Merchant, error) {
	var merchant models.Merchant
	if err := r.db.
		Preload("Inventory").
		First(&merchant, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &merchant, nil
}

// Trova esercenti nelle vicinanze
func (r *MerchantRepository) GetNearby(latitude, longitude, radiusKm float64, merchantType *models.MerchantType) ([]models.Merchant, error) {
	var merchants []models.Merchant
	query := r.db.
		Where("status = ?", models.MerchantStatusActive).
		Where(
			"(6371 * acos(cos(radians(?)) * cos(radians(latitude)) * cos(radians(longitude) - radians(?)) + sin(radians(?)) * sin(radians(latitude)))) <= ?",
			latitude, longitude, latitude, radiusKm,
		)

	if merchantType != nil {
		query = query.Where("type = ?", *merchantType)
	}

	err := query.Find(&merchants).Error
	return merchants, err
}

func (r *MerchantRepository) Update(merchant *models.Merchant) error {
	return r.db.Save(merchant).Error
}

func (r *MerchantRepository) UpdateLastActive(id uint) error {
	return r.db.Model(&models.Merchant{}).Where("id = ?", id).Update("last_active_at", time.Now()).Error
}

func (r *MerchantRepository) CountNearby(latitude, longitude, radiusKm float64) (int64, error) {
	var count int64
	err := r.db.
		Model(&models.Merchant{}).
		Where("status = ?", models.MerchantStatusActive).
		Where(
			"(6371 * acos(cos(radians(?)) * cos(radians(latitude)) * cos(radians(longitude) - radians(?)) + sin(radians(?)) * sin(radians(latitude)))) <= ?",
			latitude, longitude, latitude, radiusKm,
		).
		Count(&count).Error
	return count, err
}
