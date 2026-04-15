package models

import (
	"time"
)

type RequestMerchant struct {
	RequestID   uint      `gorm:"primaryKey" json:"request_id"`
	MerchantID  uint      `gorm:"primaryKey" json:"merchant_id"`
	NotifiedAt  time.Time `json:"notified_at"`
	RespondedAt *time.Time `json:"responded_at"`
	Response    string    `json:"response"`
	CreatedAt   time.Time `json:"created_at"`
}

func (rm *RequestMerchant) TableName() string {
	return "request_merchants"
}
