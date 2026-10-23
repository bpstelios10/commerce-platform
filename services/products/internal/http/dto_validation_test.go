package http

import (
	"context"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:fix inline
func intPtr(v int) *int {
	return new(v)
}

func TestValidateCreateProduct(t *testing.T) {
	tests := []struct {
		name                 string
		request              CreateProductRequest
		expectError          bool
		numberOfErrors       int
		expectedErrorMessage string
	}{
		{
			name: "valid product",
			request: CreateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Price:    2500,
			},
			expectError: false,
		},
		{
			name: "missing name",
			request: CreateProductRequest{
				Category: "ACCESSORY",
				Price:    2500,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "name cannot be blank or exceed 100 characters.",
		},
		{
			name: "empty name",
			request: CreateProductRequest{
				Name:     "",
				Category: "ACCESSORY",
				Price:    2500,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "name cannot be blank or exceed 100 characters.",
		},
		{
			name: "blank name",
			request: CreateProductRequest{
				Name:     "   ",
				Category: "ACCESSORY",
				Price:    2500,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "name cannot be blank or exceed 100 characters.",
		},
		{
			name: "long name",
			request: CreateProductRequest{
				Name:     strings.Repeat("a", 101),
				Category: "ACCESSORY",
				Price:    2500,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "name cannot be blank or exceed 100 characters.",
		},
		{
			name: "long name, special chars",
			request: CreateProductRequest{
				Name:     strings.Repeat("έ", 99),
				Category: "ACCESSORY",
				Price:    2500,
			},
			expectError: false,
		},
		{
			name: "missing category",
			request: CreateProductRequest{
				Name:  "MacBook Pro",
				Price: 2500,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "category cannot be blank or exceed 50 characters.",
		},
		{
			name: "empty category",
			request: CreateProductRequest{
				Name:     "MacBook Pro",
				Category: "",
				Price:    2500,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "category cannot be blank or exceed 50 characters.",
		},
		{
			name: "blank category",
			request: CreateProductRequest{
				Name:     "MacBook Pro",
				Category: "   ",
				Price:    2500,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "category cannot be blank or exceed 50 characters.",
		},
		{
			name: "long category",
			request: CreateProductRequest{
				Name:     "MacBook Pro",
				Category: strings.Repeat("a", 51),
				Price:    2500,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "category cannot be blank or exceed 50 characters.",
		},
		{
			name: "missing price",
			request: CreateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "price must be > 0 and have at most two decimal places.",
		},
		{
			name: "negative price",
			request: CreateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Price:    -100,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "price must be > 0 and have at most two decimal places.",
		},
		{
			name: "zero price",
			request: CreateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Price:    0,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "price must be > 0 and have at most two decimal places.",
		},
		{
			name: "wrong format price",
			request: CreateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Price:    10.123,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "price must be > 0 and have at most two decimal places.",
		},
		{
			name: "missing name, missing category, zero price",
			request: CreateProductRequest{
				Name:  "",
				Price: 0,
			},
			expectError:          true,
			numberOfErrors:       3,
			expectedErrorMessage: "name cannot be blank or exceed 100 characters.; category cannot be blank or exceed 50 characters.; price must be > 0 and have at most two decimal places.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := validateCreateProduct(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.IsType(t, ValidationError{}, err)
				assert.Len(t, err.(ValidationError).Errors, tt.numberOfErrors)
				if tt.numberOfErrors > 0 {
					assert.EqualError(t, err, tt.expectedErrorMessage)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateUpdateProduct(t *testing.T) {
	tests := []struct {
		name                 string
		request              UpdateProductRequest
		expectError          bool
		numberOfErrors       int
		expectedErrorMessage string
	}{
		{
			name: "valid product",
			request: UpdateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Price:    2500,
				Version:  1,
			},
			expectError: false,
		},
		{
			name: "missing name",
			request: UpdateProductRequest{
				Category: "ACCESSORY",
				Price:    2500,
				Version:  1,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "name cannot be blank or exceed 100 characters.",
		},
		{
			name: "empty name",
			request: UpdateProductRequest{
				Name:     "",
				Category: "ACCESSORY",
				Price:    2500,
				Version:  1,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "name cannot be blank or exceed 100 characters.",
		},
		{
			name: "blank name",
			request: UpdateProductRequest{
				Name:     "   ",
				Category: "ACCESSORY",
				Price:    2500,
				Version:  1,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "name cannot be blank or exceed 100 characters.",
		},
		{
			name: "long name",
			request: UpdateProductRequest{
				Name:     strings.Repeat("a", 101),
				Category: "ACCESSORY",
				Price:    2500,
				Version:  1,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "name cannot be blank or exceed 100 characters.",
		},
		{
			name: "missing category",
			request: UpdateProductRequest{
				Name:    "MacBook Pro",
				Price:   2500,
				Version: 1,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "category cannot be blank or exceed 50 characters.",
		},
		{
			name: "empty category",
			request: UpdateProductRequest{
				Name:     "MacBook Pro",
				Category: "",
				Price:    2500,
				Version:  1,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "category cannot be blank or exceed 50 characters.",
		},
		{
			name: "blank category",
			request: UpdateProductRequest{
				Name:     "MacBook Pro",
				Category: "   ",
				Price:    2500,
				Version:  1,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "category cannot be blank or exceed 50 characters.",
		},
		{
			name: "long category",
			request: UpdateProductRequest{
				Name:     "MacBook Pro",
				Category: strings.Repeat("a", 51),
				Price:    2500,
				Version:  1,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "category cannot be blank or exceed 50 characters.",
		},
		{
			name: "missing price",
			request: UpdateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Version:  1,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "price must be > 0 and have at most two decimal places.",
		},
		{
			name: "negative price",
			request: UpdateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Price:    -100,
				Version:  1,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "price must be > 0 and have at most two decimal places.",
		},
		{
			name: "zero price",
			request: UpdateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Price:    0,
				Version:  1,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "price must be > 0 and have at most two decimal places.",
		},
		{
			name: "wrong format price",
			request: UpdateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Price:    10.123,
				Version:  1,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "price must be > 0 and have at most two decimal places.",
		},
		{
			name: "missing version",
			request: UpdateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Price:    100,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "version must be > 0.",
		},
		{
			name: "negative version",
			request: UpdateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Price:    100,
				Version:  -5,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "version must be > 0.",
		},
		{
			name: "zero quantity",
			request: UpdateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Price:    100,
				Version:  0,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "version must be > 0.",
		},
		{
			name: "missing name, missing category, zero price",
			request: UpdateProductRequest{
				Name:    "",
				Price:   0,
				Version: -1,
			},
			expectError:          true,
			numberOfErrors:       4,
			expectedErrorMessage: "name cannot be blank or exceed 100 characters.; category cannot be blank or exceed 50 characters.; price must be > 0 and have at most two decimal places.; version must be > 0.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := validateUpdateProduct(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.IsType(t, ValidationError{}, err)
				assert.Len(t, err.(ValidationError).Errors, tt.numberOfErrors)
				if tt.numberOfErrors > 0 {
					assert.EqualError(t, err, tt.expectedErrorMessage)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsPriceValid(t *testing.T) {
	tests := []struct {
		name  string
		price float64
		valid bool
	}{
		{"zero", 0, false},
		{"minus", -1, false},
		{"two decimals", 19.99, true},
		{"maximum", 99_999_999.99, true},
		{"above maximum", 100_000_000, false},
		{"three decimals", 10.999, false},
		{"rounds up at third decimal", 1.005, false},
		{"negative", -1, false},
		{"NaN", math.NaN(), false},
		{"positive infinity", math.Inf(1), false},
		{"negative infinity", math.Inf(-1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, isPriceValid(tt.price))
		})
	}
}
