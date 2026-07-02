package httpx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			expectedErrorMessage: "name cannot be blank.",
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
			expectedErrorMessage: "name cannot be blank.",
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
			expectedErrorMessage: "name cannot be blank.",
		},
		{
			name: "missing price",
			request: CreateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
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
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "price must be > 0.",
		},
		{
			name: "missing name, zero price",
			request: CreateProductRequest{
				Name:     "",
				Category: "ACCESSORY",
				Price:    0,
			},
			expectError:          true,
			numberOfErrors:       2,
			expectedErrorMessage: "name cannot be blank.; price must be > 0.",
		},
		{
			name: "missing category",
			request: CreateProductRequest{
				Name:  "MacBook Pro",
				Price: 2500,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "category cannot be blank.",
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
			},
			expectError: false,
		},
		{
			name: "missing name",
			request: UpdateProductRequest{
				Category: "ACCESSORY",
				Price:    2500,
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
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "name cannot be blank.",
		},
		{
			name: "missing price",
			request: UpdateProductRequest{
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
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
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "price must be > 0.",
		},
		{
			name: "missing name, zero price",
			request: UpdateProductRequest{
				Name:     "",
				Category: "ACCESSORY",
				Price:    0,
			},
			expectError:          true,
			numberOfErrors:       2,
			expectedErrorMessage: "name cannot be blank.; price must be > 0.",
		},
		{
			name: "missing category",
			request: UpdateProductRequest{
				Name:  "MacBook Pro",
				Price: 2500,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "category cannot be blank.",
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
