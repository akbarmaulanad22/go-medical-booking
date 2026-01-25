package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Product struct {
	ID          uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string          `gorm:"type:varchar(255);not null"`
	Description string          `gorm:"type:text"`
	Price       decimal.Decimal `gorm:"type:decimal(10,2);not null"`
	Stock       int             `gorm:"default:0"`
	CreatedAt   time.Time       `gorm:"autoCreateTime"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime"`
}

func (Product) TableName() string {
	return "products"
}
