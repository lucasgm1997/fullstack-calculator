// Package dto defines the JSON transport contract of the API.
package dto

// CalculateRequest is the body of POST /api/calculate.
// Operands are pointers so a missing field can be distinguished from 0.
type CalculateRequest struct {
	Operation string   `json:"operation"`
	A         *float64 `json:"a"`
	B         *float64 `json:"b,omitempty"`
}

// CalculateResponse is returned on success.
type CalculateResponse struct {
	Result float64 `json:"result"`
}

// ErrorResponse is returned for every error.
type ErrorResponse struct {
	Error string `json:"error"`
}

// HealthResponse is returned by GET /api/health.
type HealthResponse struct {
	Status string `json:"status"`
}
