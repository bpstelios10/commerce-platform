package httpx

import (
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
				Stock:    intPtr(10),
			},
			expectError: false,
		},
		{
			name: "missing name",
			request: CreateProductRequest{
				Category: "ACCESSORY",
				Price:    2500,
				Stock:    intPtr(10),
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "name cannot be blank.",
		},
		{
			name: "empty name",
			request: CreateProductRequest{
				Name:     "",
				Category: "ACCESSORY",
				Price:    2500,
				Stock:    intPtr(10),
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "name cannot be blank.",
		},
		{
			name: "blank name",
			request: CreateProductRequest{
				Name:     "   ",
				Category: "ACCESSORY",
				Price:    2500,
				Stock:    intPtr(10),
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "name cannot be blank.",
		},
		{
			name: "missing category",
			request: CreateProductRequest{
				Name:  "MacBook Pro",
				Price: 2500,
				Stock: intPtr(10),
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "category cannot be blank.",
		},
		{
			name: "empty category",
			request: CreateProductRequest{
				Name:     "MacBook Pro",
				Category: "",
				Price:    2500,
				Stock:    intPtr(10),
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "category cannot be blank.",
		},
		{
			name: "blank category",
			request: CreateProductRequest{
				Name:     "MacBook Pro",
				Category: "   ",
				Price:    2500,
				Stock:    intPtr(10),
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "category cannot be blank.",
		},
		{
			name: "missing price",
			request: CreateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Stock:    intPtr(10),
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "price must be > 0.",
		},
		{
			name: "negative price",
			request: CreateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Price:    -100,
				Stock:    intPtr(10),
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "price must be > 0.",
		},
		{
			name: "zero price",
			request: CreateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Price:    0,
				Stock:    intPtr(10),
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "price must be > 0.",
		},
		{
			name: "missing stock",
			request: CreateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Price:    2500,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "stock cannot be negative.",
		},
		{
			name: "negative stock",
			request: CreateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Price:    2500,
				Stock:    intPtr(-10),
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "stock cannot be negative.",
		},
		{
			name: "zero stock",
			request: CreateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Price:    2500,
				Stock:    intPtr(0),
			},
			expectError: false,
		},
		{
			name: "missing name, missing category, zero price, negative stock",
			request: CreateProductRequest{
				Name:  "",
				Price: 0,
				Stock: intPtr(-10),
			},
			expectError:          true,
			numberOfErrors:       4,
			expectedErrorMessage: "name cannot be blank.; category cannot be blank.; price must be > 0.; stock cannot be negative.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := validateCreateProduct(tt.request)

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
				Stock:    intPtr(10),
			},
			expectError: false,
		},
		{
			name: "missing name",
			request: UpdateProductRequest{
				Category: "ACCESSORY",
				Price:    2500,
				Stock:    intPtr(10),
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "name cannot be blank.",
		},
		{
			name: "empty name",
			request: UpdateProductRequest{
				Name:     "",
				Category: "ACCESSORY",
				Price:    2500,
				Stock:    intPtr(10),
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "name cannot be blank.",
		},
		{
			name: "blank name",
			request: UpdateProductRequest{
				Name:     "   ",
				Category: "ACCESSORY",
				Price:    2500,
				Stock:    intPtr(10),
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "name cannot be blank.",
		},
		{
			name: "missing category",
			request: UpdateProductRequest{
				Name:  "MacBook Pro",
				Price: 2500,
				Stock: intPtr(10),
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "category cannot be blank.",
		},
		{
			name: "empty category",
			request: UpdateProductRequest{
				Name:     "MacBook Pro",
				Category: "",
				Price:    2500,
				Stock:    intPtr(10),
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "category cannot be blank.",
		},
		{
			name: "blank category",
			request: UpdateProductRequest{
				Name:     "MacBook Pro",
				Category: "   ",
				Price:    2500,
				Stock:    intPtr(10),
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "category cannot be blank.",
		},
		{
			name: "missing price",
			request: UpdateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Stock:    intPtr(10),
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "price must be > 0.",
		},
		{
			name: "negative price",
			request: UpdateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Price:    -100,
				Stock:    intPtr(10),
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "price must be > 0.",
		},
		{
			name: "zero price",
			request: UpdateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Price:    0,
				Stock:    intPtr(10),
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "price must be > 0.",
		},
		{
			name: "missing stock",
			request: UpdateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Price:    2500,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "stock cannot be negative.",
		},
		{
			name: "negative stock",
			request: UpdateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Price:    2500,
				Stock:    intPtr(-10),
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "stock cannot be negative.",
		},
		{
			name: "zero stock",
			request: UpdateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Price:    2500,
				Stock:    intPtr(0),
			},
			expectError: false,
		},
		{
			name: "missing name, missing category, zero price",
			request: UpdateProductRequest{
				Name:  "",
				Price: 0,
			},
			expectError:          true,
			numberOfErrors:       4,
			expectedErrorMessage: "name cannot be blank.; category cannot be blank.; price must be > 0.; stock cannot be negative.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := validateUpdateProduct(tt.request)

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
