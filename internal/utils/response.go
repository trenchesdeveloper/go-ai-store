package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)


type Response struct {
	Data interface{} `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
	Success bool `json:"success"`
	Message string `json:"message,omitempty"`
}

type PaginatedResponse struct {
	Response
	Meta PaginationMeta `json:"meta,omitempty"`
}

type PaginationMeta struct {
	Page int `json:"page"`
	Limit int `json:"limit"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

func SuccessResponse(c *gin.Context,message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Data:    data,
		Success: true,
		Message: message,
	})
}

func CreatedResponse(c *gin.Context,message string, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Data:    data,
		Success: true,
		Message: message,
	})
}

func ErrorResponse(c *gin.Context,message string, statusCode int, err error) {
	response := Response{
		Success: false,
		Error: message,
	}

	if err != nil {
		response.Error = err.Error()
	}

	c.JSON(statusCode, response)
}


func BadRequestResponse(c *gin.Context,message string, err error) {
	ErrorResponse(c, message, http.StatusBadRequest, err)
}

func NotFoundResponse(c *gin.Context,message string, err error) {
	ErrorResponse(c, message, http.StatusNotFound, err)
}

func ForbiddenResponse(c *gin.Context,message string, err error) {
	ErrorResponse(c, message, http.StatusForbidden, err)
}

func UnauthorizedResponse(c *gin.Context,message string, err error) {
	ErrorResponse(c, message, http.StatusUnauthorized, err)
}

func InternalErrorResponse(c *gin.Context,message string, err error) {
	ErrorResponse(c, message, http.StatusInternalServerError, err)
}

func PaginatedSuccessResponse(c *gin.Context,message string, data interface{}, meta PaginationMeta) {
	c.JSON(http.StatusOK, PaginatedResponse{
		Response: Response{
			Data:    data,
			Success: true,
			Message: message,
		},
		Meta: meta,
	})
}