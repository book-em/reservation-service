package internal

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return e.Message
}

func ErrNotFound(resourceName string, id uint) *APIError {
	return &APIError{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("Resource %s with id %d not found", resourceName, id),
	}
}

func ErrBadRequestCustom(msg string) *APIError {
	return &APIError{
		Code:    http.StatusBadRequest,
		Message: msg,
	}
}

var (
	ErrUnauthorized    = &APIError{Code: http.StatusUnauthorized, Message: "Unauthorized"}
	ErrBadRequest      = &APIError{Code: http.StatusBadRequest, Message: "Bad request"}
	ErrUnauthenticated = &APIError{Code: http.StatusForbidden, Message: "Unauthenticated"}
	ErrConflict        = &APIError{Code: http.StatusConflict, Message: "Conflict"}
)

func MapErrorToHTTP(err error) (int, string) {
	switch e := err.(type) {
	case *APIError:
		return e.Code, e.Message
	default:
		return http.StatusInternalServerError, fmt.Sprintf("Internal server error: %v", err)
	}
}

func AbortError(c *gin.Context, err error) {
	status, message := MapErrorToHTTP(err)
	c.AbortWithStatusJSON(status, gin.H{
		"error": message,
	})
	log.Printf("[ERROR] %s, returning HTTP %d, %v", message, status, err)
}
