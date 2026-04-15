package repository

import (
	"marketplace-platform/models"

	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(notification *models.MerchantNotification) error {
	return r.db.Create(notification).Error
}

func (r *NotificationRepository) GetByMerchantID(merchantID uint) ([]models.MerchantNotification, error) {
	var notifications []models.MerchantNotification
	err := r.db.
		Where("merchant_id = ?", merchantID).
		Preload("Request").
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

func (r *NotificationRepository) GetUnread(merchantID uint) ([]models.MerchantNotification, error) {
	var notifications []models.MerchantNotification
	err := r.db.
		Where("merchant_id = ? AND is_read = ?", merchantID, false).
		Preload("Request").
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

func (r *NotificationRepository) MarkAsRead(id uint) error {
	return r.db.Model(&models.MerchantNotification{}).Where("id = ?", id).Update("is_read", true).Error
}

func (r *NotificationRepository) MarkAllAsRead(merchantID uint) error {
	return r.db.Model(&models.MerchantNotification{}).Where("merchant_id = ?", merchantID).Update("is_read", true).Error
}
