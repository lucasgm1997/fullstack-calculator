// Package usecase orchestrates calculator requests. It validates the input
// shape required by each operation and delegates the arithmetic to the domain.
package usecase

import (
	"errors"
	"fmt"

	"calculator-api/internal/domain"
)

// ErrMissingOperand is returned when a required operand is not provided.
var ErrMissingOperand = errors.New("missing operand")

// CalculateInput is the use-case input. Operands are pointers so that a
// missing value can be told apart from an explicit zero.
type CalculateInput struct {
	Operation string
	A         *float64
	B         *float64
}

// Calculator executes calculations.
type Calculator struct{}

// NewCalculator creates a Calculator.
func NewCalculator() *Calculator {
	return &Calculator{}
}

// Execute validates the input and returns the calculation result.
func (c *Calculator) Execute(in CalculateInput) (float64, error) {
	op := domain.Operation(in.Operation)
	if !op.IsValid() {
		return 0, fmt.Errorf("%w %q", domain.ErrUnsupportedOperation, in.Operation)
	}
	if in.A == nil {
		return 0, fmt.Errorf("%w: field \"a\" is required", ErrMissingOperand)
	}

	var b float64
	if !op.IsUnary() {
		if in.B == nil {
			return 0, fmt.Errorf("%w: field \"b\" is required for %s", ErrMissingOperand, op)
		}
		b = *in.B
	}

	return domain.Calculate(op, *in.A, b)
}
