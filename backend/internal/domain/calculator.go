// Package domain contains the pure arithmetic rules of the calculator.
// It has no knowledge of HTTP, JSON or any framework.
package domain

import (
	"errors"
	"math"
)

// Operation identifies an arithmetic operation.
type Operation string

const (
	Add        Operation = "add"
	Subtract   Operation = "subtract"
	Multiply   Operation = "multiply"
	Divide     Operation = "divide"
	Power      Operation = "power"
	Sqrt       Operation = "sqrt"
	Percentage Operation = "percentage"
)

// Domain errors. They are sentinel values so callers can match them with errors.Is.
var (
	ErrUnsupportedOperation = errors.New("unsupported operation")
	ErrDivisionByZero       = errors.New("division by zero")
	ErrNegativeSqrt         = errors.New("square root of a negative number")
	ErrNonFiniteResult      = errors.New("result is not a finite number")
)

// IsValid reports whether op is a known operation.
func (op Operation) IsValid() bool {
	switch op {
	case Add, Subtract, Multiply, Divide, Power, Sqrt, Percentage:
		return true
	}
	return false
}

// IsUnary reports whether op takes a single operand. Unary operations ignore b.
func (op Operation) IsUnary() bool {
	return op == Sqrt
}

// Calculate applies op to the operands. For unary operations b is ignored.
//
// Percentage is defined as "a percent of b": (a / 100) * b.
func Calculate(op Operation, a, b float64) (float64, error) {
	var result float64

	switch op {
	case Add:
		result = a + b
	case Subtract:
		result = a - b
	case Multiply:
		result = a * b
	case Divide:
		if b == 0 {
			return 0, ErrDivisionByZero
		}
		result = a / b
	case Power:
		result = math.Pow(a, b)
	case Sqrt:
		if a < 0 {
			return 0, ErrNegativeSqrt
		}
		result = math.Sqrt(a)
	case Percentage:
		result = a / 100 * b
	default:
		return 0, ErrUnsupportedOperation
	}

	// Overflow (e.g. 1e308 * 10) or undefined results (e.g. (-8)^0.5) cannot be
	// represented in JSON and are not meaningful calculator output.
	if math.IsInf(result, 0) || math.IsNaN(result) {
		return 0, ErrNonFiniteResult
	}
	return result, nil
}
