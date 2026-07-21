package http

import (
	"commerce-platform/services/orders/internal/order"
	"context"
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
				ProductID: "1",
				Quantity:  10,
			},
			expectError: false,
		},
		{
			name: "missing product id",
			request: CreateOrderRequest{
				Quantity: 10,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "product-id cannot be blank.",
		},
		{
			name: "empty product id",
			request: CreateOrderRequest{
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
				ProductID: "1",
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "quantity must be > 0.",
		},
		{
			name: "negative quantity",
			request: CreateOrderRequest{
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
				ProductID: "1",
				Quantity:  0,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "quantity must be > 0.",
		},
		{
			name: "missing product id, zero quantity",
			request: CreateOrderRequest{
				ProductID: "",
				Quantity:  0,
			},
			expectError:          true,
			numberOfErrors:       2,
			expectedErrorMessage: "product-id cannot be blank.; quantity must be > 0.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := validateCreateOrder(context.Background(), tt.request)

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

func TestValidateUpdateOrder(t *testing.T) {
	tests := []struct {
		name                 string
		request              UpdateOrderRequest
		expectError          bool
		numberOfErrors       int
		expectedErrorMessage string
	}{
		{
			name: "valid Order",
			request: UpdateOrderRequest{
				ProductID: "2",
				Quantity:  5,
				Status:    order.PAID,
			},
			expectError: false,
		},
		{
			name: "missing ProductID",
			request: UpdateOrderRequest{
				Quantity: 5,
				Status:   order.PAID,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "product-id cannot be blank.",
		},
		{
			name: "empty ProductId",
			request: UpdateOrderRequest{
				ProductID: "",
				Quantity:  5,
				Status:    order.PAID,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "product-id cannot be blank.",
		},
		{
			name: "blank ProductId",
			request: UpdateOrderRequest{
				ProductID: "   ",
				Quantity:  5,
				Status:    order.PAID,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "product-id cannot be blank.",
		},
		{
			name: "missing quantity",
			request: UpdateOrderRequest{
				ProductID: "2",
				Status:    order.PAID,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "quantity must be > 0.",
		},
		{
			name: "negative quantity",
			request: UpdateOrderRequest{
				ProductID: "2",
				Quantity:  -5,
				Status:    order.PAID,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "quantity must be > 0.",
		},
		{
			name: "zero quantity",
			request: UpdateOrderRequest{
				ProductID: "2",
				Quantity:  0,
				Status:    order.PAID,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "quantity must be > 0.",
		},
		{
			name: "missing status",
			request: UpdateOrderRequest{
				ProductID: "2",
				Quantity:  1,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "status is not valid.",
		},
		{
			name: "empty status",
			request: UpdateOrderRequest{
				ProductID: "2",
				Quantity:  1,
				Status:    order.OrderStatus(" "),
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "status is not valid.",
		},
		{
			name: "invalid status",
			request: UpdateOrderRequest{
				ProductID: "2",
				Quantity:  1,
				Status:    order.OrderStatus("PIAD"),
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "status is not valid.",
		},
		{
			name: "invalid status with lowercase status",
			request: UpdateOrderRequest{
				ProductID: "2",
				Quantity:  5,
				Status:    order.OrderStatus("paid"),
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "status is not valid.",
		},
		{
			name: "empty name, zero price, invalid status",
			request: UpdateOrderRequest{
				ProductID: "",
				Quantity:  -2,
				Status:    order.OrderStatus("PIAD"),
			},
			expectError:          true,
			numberOfErrors:       3,
			expectedErrorMessage: "product-id cannot be blank.; quantity must be > 0.; status is not valid.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := validateUpdateOrder(context.Background(), tt.request)

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
