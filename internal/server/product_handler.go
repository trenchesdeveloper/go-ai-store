package server

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trenchesdeveloper/go-ai-store/internal/dto"
	"github.com/trenchesdeveloper/go-ai-store/internal/services"
	"github.com/trenchesdeveloper/go-ai-store/internal/utils"
)

func (s *Server) CreateCategory(ctx *gin.Context) {
	var req dto.CreateCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(ctx, "Invalid request payload", err)
		return
	}

	categoryService := services.NewProductService(s.store)
	category, err := categoryService.CreateCategory(ctx, req)
	if err != nil {
		utils.InternalErrorResponse(ctx, "Failed to create category", err)
		return
	}

	utils.SuccessResponse(ctx, "Category created successfully", category)
}

func (s *Server) GetCategories(ctx *gin.Context) {
	categoriesService := services.NewProductService(s.store)
	categories, err := categoriesService.GetCategories(ctx)
	if err != nil {
		utils.InternalErrorResponse(ctx, "Failed to get categories", err)
		return
	}

	utils.SuccessResponse(ctx, "Categories retrieved successfully", categories)
}

func (s *Server) UpdateCategory(ctx *gin.Context) {
	var req dto.UpdateCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(ctx, "Invalid request payload", err)
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequestResponse(ctx, "Invalid category ID", err)
		return
	}

	req.ID = int32(id) //#nosec G115 -- category ID is bounded by database constraints

	categoryService := services.NewProductService(s.store)
	category, err := categoryService.UpdateCategory(ctx, req)
	if err != nil {
		utils.InternalErrorResponse(ctx, "Failed to update category", err)
		return
	}

	utils.SuccessResponse(ctx, "Category updated successfully", category)
}

func (s *Server) DeleteCategory(ctx *gin.Context) {
	categoryService := services.NewProductService(s.store)
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequestResponse(ctx, "Invalid category ID", err)
		return
	}

	err = categoryService.DeleteCategory(ctx, uint(id))
	if err != nil {
		utils.InternalErrorResponse(ctx, "Failed to delete category", err)
		return
	}

	utils.SuccessResponse(ctx, "Category deleted successfully", nil)
}

func (s *Server) CreateProduct(ctx *gin.Context) {
	var req dto.CreateProductRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(ctx, "Invalid request payload", err)
		return
	}

	productService := services.NewProductService(s.store)
	product, err := productService.CreateProduct(ctx, req)
	if err != nil {
		utils.InternalErrorResponse(ctx, "Failed to create product", err)
		return
	}

	utils.SuccessResponse(ctx, "Product created successfully", product)
}

func (s *Server) GetProducts(ctx *gin.Context) {
	// Parse pagination query parameters
	page := 1
	limit := 10

	if pageStr := ctx.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	productsService := services.NewProductService(s.store)
	products, paginationMeta, err := productsService.GetProducts(ctx, page, limit)
	if err != nil {
		utils.InternalErrorResponse(ctx, "Failed to get products", err)
		return
	}

	utils.PaginatedSuccessResponse(ctx, "Products retrieved successfully", products, *paginationMeta)
}

func (s *Server) GetProductByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequestResponse(ctx, "Invalid product ID", err)
		return
	}

	productService := services.NewProductService(s.store)
	product, err := productService.GetProductByID(ctx, uint(id))
	if err != nil {
		utils.NotFoundResponse(ctx, "Product not found", err)
		return
	}

	utils.SuccessResponse(ctx, "Product retrieved successfully", product)
}

func (s *Server) UpdateProductByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequestResponse(ctx, "Invalid product ID", err)
		return
	}

	var req dto.UpdateProductRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(ctx, "Invalid request payload", err)
		return
	}

	productService := services.NewProductService(s.store)
	product, err := productService.UpdateProductByID(ctx, uint(id), &req)
	if err != nil {
		utils.InternalErrorResponse(ctx, "Failed to update product", err)
		return
	}

	utils.SuccessResponse(ctx, "Product updated successfully", product)
}

func (s *Server) DeleteProductByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequestResponse(ctx, "Invalid product ID", err)
		return
	}

	productService := services.NewProductService(s.store)
	err = productService.DeleteProductByID(ctx, uint(id))
	if err != nil {
		utils.InternalErrorResponse(ctx, "Failed to delete product", err)
		return
	}

	utils.SuccessResponse(ctx, "Product deleted successfully", nil)
}
