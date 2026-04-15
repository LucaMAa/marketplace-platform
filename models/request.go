package models

import (
	"time"

	"gorm.io/datatypes"
)

type RequestStatus string

const (
	RequestPending    RequestStatus = "pending"
	RequestMatched    RequestStatus = "matched"
	RequestCompleted  RequestStatus = "completed"
	RequestCancelled  RequestStatus = "cancelled"
)

type Request struct {
	ID            uint              `gorm:"primaryKey" json:"id"`
	UserID        uint              `json:"user_id"`
	User          *User             `gorm:"foreignKey:UserID" json:"-"`
	ProductCode   string            `json:"product_code"`
	ProductName   string            `json:"product_name"`
	Quantity      int               `json:"quantity"`
	Status        RequestStatus     `json:"status"`
	Latitude      float64           `json:"latitude"`
	Longitude     float64           `json:"longitude"`
	Radius        int               `json:"radius"`
	Merchants     []Merchant        `gorm:"many2many:request_merchants;" json:"-"`
	Messages      []ChatMessage     `gorm:"foreignKey:RequestID" json:"-"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	ExpiresAt     time.Time         `json:"expires_at"`
	Metadata      datatypes.JSONMap `gorm:"type:jsonb" json:"metadata"`
}

func (r *Request) TableName() string {
	return "requests"
}
