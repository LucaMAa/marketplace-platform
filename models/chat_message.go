package models

import (
	"time"
)

type MessageType string

const (
	MessageTypeRequest  MessageType = "request"
	MessageTypeResponse MessageType = "response"
	MessageTypeChat     MessageType = "chat"
)

type ChatMessage struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	RequestID uint        `json:"request_id"`
	Request   *Request    `gorm:"foreignKey:RequestID" json:"-"`
	UserID    *uint       `json:"user_id"`
	User      *User       `gorm:"foreignKey:UserID" json:"-"`
	MerchantID *uint      `json:"merchant_id"`
	Merchant  *Merchant   `gorm:"foreignKey:MerchantID" json:"-"`
	Type      MessageType `json:"type"`
	Content   string      `json:"content"`
	HasProduct bool       `json:"has_product"`
	CanOrder  bool        `json:"can_order"`
	Price     *float64    `json:"price"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

func (m *ChatMessage) TableName() string {
	return "chat_messages"
}
