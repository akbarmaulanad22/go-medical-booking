package repository

import (
	"context"

	"go-template-clean-architecture/internal/domain/entity"

	"github.com/google/uuid"
)

type ProductRepository interface {
	Create(ctx context.Context, product *entity.Product) error
	FindAll(ctx context.Context, limit, offset int) ([]entity.Product, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Product, error)
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
}
