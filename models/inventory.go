package models

import (
	"time"
)

type Inventory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MerchantID uint     `json:"merchant_id"`
	Merchant  *Merchant `gorm:"foreignKey:MerchantID" json:"-"`
	ProductID uint      `json:"product_id"`
	Product   *Product  `gorm:"foreignKey:ProductID" json:"-"`
	Quantity  int       `json:"quantity"`
	CanOrder  bool      `json:"can_order"`
	Price     *float64  `json:"price"`
	LastUpdate time.Time `json:"last_update"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (i *Inventory) TableName() string {
	return "inventories"
}
