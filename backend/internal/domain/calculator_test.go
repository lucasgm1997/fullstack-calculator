package domain

import (
	"errors"
	"math"
	"testing"
)

func TestCalculate(t *testing.T) {
	tests := []struct {
		name string
		op   Operation
		a, b float64
		want float64
	}{
		{"add", Add, 10, 5, 15},
		{"add negatives", Add, -2.5, -2.5, -5},
		{"subtract", Subtract, 10, 5, 5},
		{"subtract to negative", Subtract, 5, 10, -5},
		{"multiply", Multiply, 10, 5, 50},
		{"multiply by zero", Multiply, 10, 0, 0},
		{"divide", Divide, 10, 4, 2.5},
		{"divide negative", Divide, -9, 3, -3},
		{"divide zero by number", Divide, 0, 5, 0},
		{"power", Power, 2, 10, 1024},
		{"power negative exponent", Power, 2, -1, 0.5},
		{"power zero exponent", Power, 0, 0, 1},
		{"sqrt", Sqrt, 16, 0, 4},
		{"sqrt ignores b", Sqrt, 9, 123, 3},
		{"sqrt zero", Sqrt, 0, 0, 0},
		{"percentage", Percentage, 20, 50, 10},
		{"percentage over 100", Percentage, 150, 20, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Calculate(tt.op, tt.a, tt.b)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("Calculate(%s, %v, %v) = %v, want %v", tt.op, tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestCalculateErrors(t *testing.T) {
	tests := []struct {
		name    string
		op      Operation
		a, b    float64
		wantErr error
	}{
		{"division by zero", Divide, 10, 0, ErrDivisionByZero},
		{"zero divided by zero", Divide, 0, 0, ErrDivisionByZero},
		{"negative square root", Sqrt, -4, 0, ErrNegativeSqrt},
		{"overflow", Multiply, math.MaxFloat64, 10, ErrNonFiniteResult},
		{"power overflow", Power, 10, 400, ErrNonFiniteResult},
		{"undefined power", Power, -8, 0.5, ErrNonFiniteResult},
		{"unknown operation", Operation("modulo"), 1, 2, ErrUnsupportedOperation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Calculate(tt.op, tt.a, tt.b)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestOperationIsValid(t *testing.T) {
	for _, op := range Operations {
		if !op.IsValid() {
			t.Errorf("%q should be valid", op)
		}
	}
	for _, op := range []Operation{"", "ADD", "modulo"} {
		if op.IsValid() {
			t.Errorf("%q should be invalid", op)
		}
	}
}

func TestOperationIsUnary(t *testing.T) {
	if !Sqrt.IsUnary() {
		t.Error("sqrt should be unary")
	}
	if Add.IsUnary() || Divide.IsUnary() || Percentage.IsUnary() {
		t.Error("binary operations should not be unary")
	}
}
