package repository

import (
	"errors"
	"marketplace-platform/models"
	"time"

	"gorm.io/gorm"
)

type RequestRepository struct {
	db *gorm.DB
}

func NewRequestRepository(db *gorm.DB) *RequestRepository {
	return &RequestRepository{db: db}
}

func (r *RequestRepository) Create(request *models.Request) error {
	return r.db.Create(request).Error
}

func (r *RequestRepository) GetByID(id uint) (*models.Request, error) {
	var request models.Request
	if err := r.db.
		Preload("Merchants").
		Preload("Messages").
		Preload("User").
		First(&request, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &request, nil
}

func (r *RequestRepository) GetByUserID(userID uint) ([]models.Request, error) {
	var requests []models.Request
	err := r.db.
		Where("user_id = ?", userID).
		Preload("Merchants").
		Preload("Messages").
		Order("created_at DESC").
		Find(&requests).Error
	return requests, err
}

func (r *RequestRepository) Update(request *models.Request) error {
	return r.db.Save(request).Error
}

func (r *RequestRepository) AddMerchant(requestID, merchantID uint) error {
	return r.db.
		Model(&models.Request{}).
		Where("id = ?", requestID).
		Association("Merchants").
		Append(&models.Merchant{ID: merchantID})
}

// Ottieni richieste scadute
func (r *RequestRepository) GetExpiredRequests() ([]models.Request, error) {
	var requests []models.Request
	err := r.db.
		Where("status = ? AND expires_at < ?", models.RequestPending, time.Now()).
		Find(&requests).Error
	return requests, err
}

// Ottieni richieste recenti per learning
func (r *RequestRepository) GetRecentRequests(limit int) ([]models.Request, error) {
	var requests []models.Request
	err := r.db.
		Order("created_at DESC").
		Limit(limit).
		Find(&requests).Error
	return requests, err
}
