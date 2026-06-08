package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCreateOrder(t *testing.T) {
	tests := []struct {
		name                 string
		request              CreateOrderRequest
		expectError          bool
		numberOfErrors       int
		expectedErrorMessage string
	}{
		{
			name: "valid Order",
			request: CreateOrderRequest{
				ID:        "1",
				ProductID: "1",
				Quantity:  10,
			},
			expectError: false,
		},
		{
			name: "missing id",
			request: CreateOrderRequest{
				ProductID: "1",
				Quantity:  10,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "id cannot be blank.",
		},
		{
			name: "empty id",
			request: CreateOrderRequest{
				ID:        "",
				ProductID: "1",
				Quantity:  10,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "id cannot be blank.",
		},
		{
			name: "blank id",
			request: CreateOrderRequest{
				ID:        "  ",
				ProductID: "1",
				Quantity:  10,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "id cannot be blank.",
		},
		{
			name: "missing product id",
			request: CreateOrderRequest{
				ID:       "1",
				Quantity: 10,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "product-id cannot be blank.",
		},
		{
			name: "empty product id",
			request: CreateOrderRequest{
				ID:        "1",
				ProductID: "",
				Quantity:  10,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "product-id cannot be blank.",
		},
		{
			name: "blank product id",
			request: CreateOrderRequest{
				ID:        "1",
				ProductID: "   ",
				Quantity:  10,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "product-id cannot be blank.",
		},
		{
			name: "missing quantity",
			request: CreateOrderRequest{
				ID:        "1",
				ProductID: "1",
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "quantity must be > 0.",
		},
		{
			name: "negative quantity",
			request: CreateOrderRequest{
				ID:        "1",
				ProductID: "1",
				Quantity:  -100,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "quantity must be > 0.",
		},
		{
			name: "zero quantity",
			request: CreateOrderRequest{
				ID:        "1",
				ProductID: "1",
				Quantity:  0,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "quantity must be > 0.",
		},
		{
			name: "missing id and product id, zero quantity",
			request: CreateOrderRequest{
				ID:        "",
				ProductID: "",
				Quantity:  0,
			},
			expectError:          true,
			numberOfErrors:       3,
			expectedErrorMessage: "id cannot be blank.; product-id cannot be blank.; quantity must be > 0.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := validateCreateOrder(tt.request)

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
