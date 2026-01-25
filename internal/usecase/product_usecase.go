package usecase

import (
	"context"
	"errors"

	"go-template-clean-architecture/internal/delivery/dto"
	"go-template-clean-architecture/internal/domain/entity"
	"go-template-clean-architecture/internal/domain/repository"

	"github.com/google/uuid"
)

var (
	ErrProductNotFound = errors.New("product not found")
)

type ProductUsecase interface {
	Create(ctx context.Context, req *dto.CreateProductRequest) (*dto.ProductResponse, error)
	GetAll(ctx context.Context, page, limit int) ([]dto.ProductResponse, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.ProductResponse, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateProductRequest) (*dto.ProductResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type productUsecase struct {
	productRepo repository.ProductRepository
}

func NewProductUsecase(productRepo repository.ProductRepository) ProductUsecase {
	return &productUsecase{productRepo: productRepo}
}

func (u *productUsecase) Create(ctx context.Context, req *dto.CreateProductRequest) (*dto.ProductResponse, error) {
	product := &entity.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	}

	if err := u.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}

	return u.toProductResponse(product), nil
}

func (u *productUsecase) GetAll(ctx context.Context, page, limit int) ([]dto.ProductResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	products, total, err := u.productRepo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	var responses []dto.ProductResponse
	for _, product := range products {
		responses = append(responses, *u.toProductResponse(&product))
	}

	return responses, total, nil
}

func (u *productUsecase) GetByID(ctx context.Context, id uuid.UUID) (*dto.ProductResponse, error) {
	product, err := u.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, ErrProductNotFound
	}

	return u.toProductResponse(product), nil
}

func (u *productUsecase) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	product, err := u.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, ErrProductNotFound
	}

	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.Stock = req.Stock

	if err := u.productRepo.Update(ctx, product); err != nil {
		return nil, err
	}

	return u.toProductResponse(product), nil
}

func (u *productUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	product, err := u.productRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if product == nil {
		return ErrProductNotFound
	}

	return u.productRepo.Delete(ctx, id)
}

func (u *productUsecase) toProductResponse(product *entity.Product) *dto.ProductResponse {
	return &dto.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}
