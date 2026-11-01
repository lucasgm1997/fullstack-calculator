package usecase

import (
	"errors"
	"testing"

	"calculator-api/internal/domain"
)

func ptr(v float64) *float64 { return &v }

func TestExecute(t *testing.T) {
	calc := NewCalculator()

	tests := []struct {
		name    string
		in      CalculateInput
		want    float64
		wantErr error
	}{
		{name: "binary operation", in: CalculateInput{Operation: "add", A: ptr(10), B: ptr(5)}, want: 15},
		{name: "explicit zero operand is valid", in: CalculateInput{Operation: "multiply", A: ptr(0), B: ptr(5)}, want: 0},
		{name: "unary operation without b", in: CalculateInput{Operation: "sqrt", A: ptr(25)}, want: 5},
		{name: "unsupported operation", in: CalculateInput{Operation: "modulo", A: ptr(1), B: ptr(2)}, wantErr: domain.ErrUnsupportedOperation},
		{name: "empty operation", in: CalculateInput{A: ptr(1), B: ptr(2)}, wantErr: domain.ErrUnsupportedOperation},
		{name: "missing a", in: CalculateInput{Operation: "add", B: ptr(2)}, wantErr: ErrMissingOperand},
		{name: "missing b for binary operation", in: CalculateInput{Operation: "divide", A: ptr(2)}, wantErr: ErrMissingOperand},
		{name: "missing a for unary operation", in: CalculateInput{Operation: "sqrt"}, wantErr: ErrMissingOperand},
		{name: "domain error is propagated", in: CalculateInput{Operation: "divide", A: ptr(1), B: ptr(0)}, wantErr: domain.ErrDivisionByZero},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calc.Execute(tt.in)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("result = %v, want %v", got, tt.want)
			}
		})
	}
}
