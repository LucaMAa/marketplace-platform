package models

import (
	"time"
)

type NotificationType string

const (
	NotificationTypeNewRequest NotificationType = "new_request"
	NotificationTypeMessage    NotificationType = "message"
)

type MerchantNotification struct {
	ID         uint             `gorm:"primaryKey" json:"id"`
	MerchantID uint             `json:"merchant_id"`
	Merchant   *Merchant        `gorm:"foreignKey:MerchantID" json:"-"`
	RequestID  uint             `json:"request_id"`
	Request    *Request         `gorm:"foreignKey:RequestID" json:"-"`
	Type       NotificationType `json:"type"`
	IsRead     bool             `json:"is_read"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

func (mn *MerchantNotification) TableName() string {
	return "merchant_notifications"
}
