package repository

import (
	"marketplace-platform/models"

	"gorm.io/gorm"
)

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) Create(message *models.ChatMessage) error {
	return r.db.Create(message).Error
}

func (r *ChatRepository) GetByRequestID(requestID uint) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	err := r.db.
		Where("request_id = ?", requestID).
		Preload("User").
		Preload("Merchant").
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}

func (r *ChatRepository) GetByRequestIDAndMerchant(requestID, merchantID uint) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	err := r.db.
		Where("request_id = ? AND merchant_id = ?", requestID, merchantID).
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}

func (r *ChatRepository) GetLatest(requestID, merchantID uint, limit int) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	err := r.db.
		Where("request_id = ? AND merchant_id = ?", requestID, merchantID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}
