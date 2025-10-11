package internal

import (
	"bookem-reservation-service/util"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
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

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "status", "endpoint"},
	)

	httpResponseSizeBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_response_size_bytes",
			Help: "Total response size in bytes",
		},
		[]string{"endpoint", "status"},
	)
)

func PrometheusMiddleware() gin.HandlerFunc {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpResponseSizeBytes)

	return func(c *gin.Context) {
		c.Next()

		endpoint := c.FullPath()
		status := fmt.Sprintf("%d", c.Writer.Status())
		method := c.Request.Method
		size := float64(c.Writer.Size())

		httpRequestsTotal.WithLabelValues(method, status, endpoint).Inc()

		if size >= 0 {
			httpResponseSizeBytes.WithLabelValues(endpoint, status).Add(float64(size))
		} else {
			util.TEL.Warn("Response size < 0, cannot push to Prometheus", "size", size)
		}
	}
}
