package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Request DTOs

type CreateProductRequest struct {
	Name        string          `json:"name" validate:"required,min=2"`
	Description string          `json:"description"`
	Price       decimal.Decimal `json:"price" validate:"required"`
	Stock       int             `json:"stock" validate:"gte=0"`
}

type UpdateProductRequest struct {
	Name        string          `json:"name" validate:"required,min=2"`
	Description string          `json:"description"`
	Price       decimal.Decimal `json:"price" validate:"required"`
	Stock       int             `json:"stock" validate:"gte=0"`
}

// Response DTOs

type ProductResponse struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Price       decimal.Decimal `json:"price"`
	Stock       int             `json:"stock"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type ProductListResponse struct {
	Products []ProductResponse `json:"products"`
}
