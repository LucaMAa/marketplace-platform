package models

import (
	"time"
)

type Product struct {
	ID          uint        `gorm:"primaryKey" json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Code        string      `gorm:"index" json:"code"`
	Category    string      `json:"category"`
	CategoryID  uint        `json:"category_id"`
	Inventories []Inventory `gorm:"foreignKey:ProductID" json:"-"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

func (p *Product) TableName() string {
	return "products"
}

type ProductCategory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex" json:"name"`
	Icon      string    `json:"icon"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (pc *ProductCategory) TableName() string {
	return "product_categories"
}
