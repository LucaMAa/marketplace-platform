package models

import (
	"time"

	"gorm.io/datatypes"
)

type MerchantType string

const (
	MerchantTypePharmacy    MerchantType = "pharmacy"
	MerchantTypeHardware    MerchantType = "hardware"
	MerchantTypeGrocery     MerchantType = "grocery"
	MerchantTypeElectronics MerchantType = "electronics"
	MerchantTypeOther       MerchantType = "other"
)

type MerchantStatus string

const (
	MerchantStatusActive   MerchantStatus = "active"
	MerchantStatusInactive MerchantStatus = "inactive"
	MerchantStatusPending  MerchantStatus = "pending"
)

type Merchant struct {
	ID              uint              `gorm:"primaryKey" json:"id"`
	Email           string            `gorm:"uniqueIndex" json:"email"`
	Password        string            `json:"-"`
	BusinessName    string            `json:"business_name"`
	Type            MerchantType      `json:"type"`
	Status          MerchantStatus    `json:"status"`
	Latitude        float64           `json:"latitude"`
	Longitude       float64           `json:"longitude"`
	Address         string            `json:"address"`
	City            string            `json:"city"`
	Phone           string            `json:"phone"`
	Website         *string           `json:"website"`
	OpenHours       datatypes.JSONMap `gorm:"type:jsonb" json:"open_hours"`
	Inventory       []Inventory       `gorm:"foreignKey:MerchantID" json:"-"`
	Messages        []ChatMessage     `gorm:"foreignKey:MerchantID" json:"-"`
	Notifications   []MerchantNotification `gorm:"foreignKey:MerchantID" json:"-"`
	Requests        []Request         `gorm:"many2many:request_merchants;" json:"-"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	LastActiveAt    *time.Time        `json:"last_active_at"`
	Metadata        datatypes.JSONMap `gorm:"type:jsonb" json:"metadata"`
}

func (m *Merchant) TableName() string {
	return "merchants"
}
