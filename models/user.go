package models

import (
	"time"

	"gorm.io/datatypes"
)

type User struct {
	ID        uint             `gorm:"primaryKey" json:"id"`
	Email     string           `gorm:"uniqueIndex" json:"email"`
	Password  string           `json:"-"`
	Name      string           `json:"name"`
	Latitude  float64          `json:"latitude"`
	Longitude float64          `json:"longitude"`
	Address   string           `json:"address"`
	Requests  []Request        `gorm:"foreignKey:UserID" json:"-"`
	Messages  []ChatMessage    `gorm:"foreignKey:UserID" json:"-"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	Metadata  datatypes.JSONMap `gorm:"type:jsonb" json:"metadata"`
}

func (u *User) TableName() string {
	return "users"
}
