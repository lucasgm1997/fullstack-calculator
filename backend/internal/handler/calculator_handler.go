// Package handler adapts HTTP requests to use-case calls and maps results
// and errors to JSON responses.
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"calculator-api/internal/domain"
	"calculator-api/internal/dto"
	"calculator-api/internal/usecase"
)

// maxBodyBytes caps the request size; a calculation payload is tiny.
const maxBodyBytes = 4 << 10

// Calculator is the use-case dependency of the handler.
type Calculator interface {
	Execute(in usecase.CalculateInput) (float64, error)
}

// CalculatorHandler serves calculator endpoints.
type CalculatorHandler struct {
	calculator Calculator
}

// NewCalculatorHandler creates a handler with its use-case dependency.
func NewCalculatorHandler(calculator Calculator) *CalculatorHandler {
	return &CalculatorHandler{calculator: calculator}
}

// Calculate handles POST /api/calculate.
func (h *CalculatorHandler) Calculate(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)

	var req dto.CalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: bindErrorMessage(err)})
		return
	}

	result, err := h.calculator.Execute(usecase.CalculateInput{
		Operation: req.Operation,
		A:         req.A,
		B:         req.B,
	})
	if err != nil {
		status, message := mapError(err)
		c.JSON(status, dto.ErrorResponse{Error: message})
		return
	}

	c.JSON(http.StatusOK, dto.CalculateResponse{Result: result})
}

// Health handles GET /api/health.
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, dto.HealthResponse{Status: "ok"})
}

// mapError converts use-case and domain errors to an HTTP status and a
// client-safe message. Unknown errors are logged and hidden behind a 500.
func mapError(err error) (int, string) {
	switch {
	case errors.Is(err, usecase.ErrMissingOperand),
		errors.Is(err, domain.ErrUnsupportedOperation),
		errors.Is(err, domain.ErrDivisionByZero),
		errors.Is(err, domain.ErrNegativeSqrt),
		errors.Is(err, domain.ErrNonFiniteResult):
		return http.StatusBadRequest, err.Error()
	default:
		log.Printf("unexpected calculation error: %v", err)
		return http.StatusInternalServerError, "internal server error"
	}
}

// bindErrorMessage turns JSON decoding errors into readable messages.
func bindErrorMessage(err error) string {
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) && typeErr.Field != "" {
		if typeErr.Field == "operation" {
			return `field "operation" must be a string`
		}
		return fmt.Sprintf("field %q must be a finite number", typeErr.Field)
	}

	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		return "request body too large"
	}

	return "request body must be valid JSON"
}
