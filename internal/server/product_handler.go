package server

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trenchesdeveloper/go-ai-store/internal/dto"
	"github.com/trenchesdeveloper/go-ai-store/internal/utils"
)

// CreateCategory godoc
// @Summary      Create category (Admin)
// @Description  Create a new product category
// @Tags         categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateCategoryRequest true "Category data"
// @Success      201  {object}  utils.Response{data=dto.CategoryResponse}
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /categories [post]
func (s *Server) CreateCategory(ctx *gin.Context) {
	var req dto.CreateCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(ctx, "Invalid request payload", err)
		return
	}

	category, err := s.productService.CreateCategory(ctx, req)
	if err != nil {
		utils.InternalErrorResponse(ctx, "Failed to create category", err)
		return
	}

	utils.CreatedResponse(ctx, "Category created successfully", category)
}

// GetCategories godoc
// @Summary      List categories
// @Description  Get all product categories
// @Tags         categories
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response{data=[]dto.CategoryResponse}
// @Failure      500  {object}  utils.Response
// @Router       /categories [get]
func (s *Server) GetCategories(ctx *gin.Context) {
	categories, err := s.productService.GetCategories(ctx)
	if err != nil {
		utils.InternalErrorResponse(ctx, "Failed to get categories", err)
		return
	}

	utils.SuccessResponse(ctx, "Categories retrieved successfully", categories)
}

// UpdateCategory godoc
// @Summary      Update category (Admin)
// @Description  Update an existing category
// @Tags         categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Category ID"
// @Param        request body dto.UpdateCategoryRequest true "Category data"
// @Success      200  {object}  utils.Response{data=dto.CategoryResponse}
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /categories/{id} [put]
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

	category, err := s.productService.UpdateCategory(ctx, req)
	if err != nil {
		utils.InternalErrorResponse(ctx, "Failed to update category", err)
		return
	}

	utils.SuccessResponse(ctx, "Category updated successfully", category)
}

// DeleteCategory godoc
// @Summary      Delete category (Admin)
// @Description  Delete a category
// @Tags         categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Category ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /categories/{id} [delete]
func (s *Server) DeleteCategory(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequestResponse(ctx, "Invalid category ID", err)
		return
	}

	err = s.productService.DeleteCategory(ctx, uint(id))
	if err != nil {
		utils.InternalErrorResponse(ctx, "Failed to delete category", err)
		return
	}

	utils.SuccessResponse(ctx, "Category deleted successfully", nil)
}

// CreateProduct godoc
// @Summary      Create product (Admin)
// @Description  Create a new product
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateProductRequest true "Product data"
// @Success      201  {object}  utils.Response{data=dto.ProductResponse}
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /products [post]
func (s *Server) CreateProduct(ctx *gin.Context) {
	var req dto.CreateProductRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(ctx, "Invalid request payload", err)
		return
	}

	product, err := s.productService.CreateProduct(ctx, req)
	if err != nil {
		utils.InternalErrorResponse(ctx, "Failed to create product", err)
		return
	}

	utils.CreatedResponse(ctx, "Product created successfully", product)
}

// GetProducts godoc
// @Summary      List products
// @Description  Get all products with pagination
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(10)
// @Success      200  {object}  utils.PaginatedResponse{data=[]dto.ProductResponse}
// @Failure      500  {object}  utils.Response
// @Router       /products [get]
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

	products, paginationMeta, err := s.productService.GetProducts(ctx, page, limit)
	if err != nil {
		utils.InternalErrorResponse(ctx, "Failed to get products", err)
		return
	}

	utils.PaginatedSuccessResponse(ctx, "Products retrieved successfully", products, *paginationMeta)
}

// GetProductByID godoc
// @Summary      Get product by ID
// @Description  Get a single product by ID
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id path int true "Product ID"
// @Success      200  {object}  utils.Response{data=dto.ProductResponse}
// @Failure      400  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /products/{id} [get]
func (s *Server) GetProductByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequestResponse(ctx, "Invalid product ID", err)
		return
	}

	product, err := s.productService.GetProductByID(ctx, uint(id))
	if err != nil {
		utils.NotFoundResponse(ctx, "Product not found", err)
		return
	}

	utils.SuccessResponse(ctx, "Product retrieved successfully", product)
}

// UpdateProductByID godoc
// @Summary      Update product (Admin)
// @Description  Update an existing product
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Product ID"
// @Param        request body dto.UpdateProductRequest true "Product data"
// @Success      200  {object}  utils.Response{data=dto.ProductResponse}
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /products/{id} [put]
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

	product, err := s.productService.UpdateProductByID(ctx, uint(id), &req)
	if err != nil {
		utils.InternalErrorResponse(ctx, "Failed to update product", err)
		return
	}

	utils.SuccessResponse(ctx, "Product updated successfully", product)
}

// DeleteProductByID godoc
// @Summary      Delete product (Admin)
// @Description  Delete a product
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Product ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /products/{id} [delete]
func (s *Server) DeleteProductByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequestResponse(ctx, "Invalid product ID", err)
		return
	}

	err = s.productService.DeleteProductByID(ctx, uint(id))
	if err != nil {
		utils.InternalErrorResponse(ctx, "Failed to delete product", err)
		return
	}

	utils.SuccessResponse(ctx, "Product deleted successfully", nil)
}

// UploadProductImage godoc
// @Summary      Upload product image (Admin)
// @Description  Upload an image for a product
// @Tags         products
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Product ID"
// @Param        image formData file true "Product image"
// @Success      200  {object}  utils.Response{data=string}
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /products/{id}/image [post]
func (s *Server) UploadProductImage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequestResponse(ctx, "Invalid product ID", err)
		return
	}

	file, err := ctx.FormFile("image")
	if err != nil {
		utils.BadRequestResponse(ctx, "Invalid file", err)
		return
	}

	imagePath, err := s.uploadService.UploadProductImage(
		uint(id),
		file,
	)
	if err != nil {
		utils.InternalErrorResponse(ctx, "Failed to upload product image", err)
		return
	}

	// Update product image path in database
	err = s.productService.UpdateProductImage(ctx, uint(id), imagePath, file.Filename)
	if err != nil {
		utils.InternalErrorResponse(ctx, "Failed to update product image", err)
		return
	}

	utils.SuccessResponse(ctx, "Product image uploaded successfully", imagePath)
}
