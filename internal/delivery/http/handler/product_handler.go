package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go-template-clean-architecture/internal/delivery/dto"
	"go-template-clean-architecture/internal/usecase"
	"go-template-clean-architecture/pkg/response"
	"go-template-clean-architecture/pkg/validator"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ProductHandler struct {
	productUsecase usecase.ProductUsecase
	validator      *validator.CustomValidator
}

func NewProductHandler(productUsecase usecase.ProductUsecase, validator *validator.CustomValidator) *ProductHandler {
	return &ProductHandler{
		productUsecase: productUsecase,
		validator:      validator,
	}
}

// Create handles product creation
// @Summary Create a new product
// @Description Create a new product
// @Tags Products
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateProductRequest true "Create Product Request"
// @Success 201 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /products [post]
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationError(w, h.validator.FormatValidationErrors(err))
		return
	}

	product, err := h.productUsecase.Create(r.Context(), &req)
	if err != nil {
		response.InternalServerError(w, "Failed to create product")
		return
	}

	response.Success(w, http.StatusCreated, "Product created successfully", product)
}

// GetAll handles getting all products
// @Summary Get all products
// @Description Get all products with pagination
// @Tags Products
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.Response
// @Router /products [get]
func (h *ProductHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	products, total, err := h.productUsecase.GetAll(r.Context(), page, limit)
	if err != nil {
		response.InternalServerError(w, "Failed to get products")
		return
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	meta := &response.Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	response.SuccessWithMeta(w, http.StatusOK, "Products retrieved successfully", products, meta)
}

// GetByID handles getting a product by ID
// @Summary Get product by ID
// @Description Get a product by its ID
// @Tags Products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /products/{id} [get]
func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid product ID", nil)
		return
	}

	product, err := h.productUsecase.GetByID(r.Context(), id)
	if err != nil {
		switch err {
		case usecase.ErrProductNotFound:
			response.NotFound(w, "Product not found")
		default:
			response.InternalServerError(w, "Failed to get product")
		}
		return
	}

	response.Success(w, http.StatusOK, "Product retrieved successfully", product)
}

// Update handles product update
// @Summary Update a product
// @Description Update a product by its ID
// @Tags Products
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param request body dto.UpdateProductRequest true "Update Product Request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /products/{id} [put]
func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid product ID", nil)
		return
	}

	var req dto.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationError(w, h.validator.FormatValidationErrors(err))
		return
	}

	product, err := h.productUsecase.Update(r.Context(), id, &req)
	if err != nil {
		switch err {
		case usecase.ErrProductNotFound:
			response.NotFound(w, "Product not found")
		default:
			response.InternalServerError(w, "Failed to update product")
		}
		return
	}

	response.Success(w, http.StatusOK, "Product updated successfully", product)
}

// Delete handles product deletion
// @Summary Delete a product
// @Description Delete a product by its ID
// @Tags Products
// @Security BearerAuth
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /products/{id} [delete]
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid product ID", nil)
		return
	}

	err = h.productUsecase.Delete(r.Context(), id)
	if err != nil {
		switch err {
		case usecase.ErrProductNotFound:
			response.NotFound(w, "Product not found")
		default:
			response.InternalServerError(w, "Failed to delete product")
		}
		return
	}

	response.Success(w, http.StatusOK, "Product deleted successfully", nil)
}
