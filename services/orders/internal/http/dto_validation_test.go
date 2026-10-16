package http

import (
	"commerce-platform/services/orders/internal/order"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCreateOrder(t *testing.T) {
	const validUUID = "f47ac10b-58cc-4372-a567-0e02b2c3d009"

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
				ProductID: validUUID,
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
			name: "product id not valid UUID",
			request: CreateOrderRequest{
				ProductID: "1",
				Quantity:  10,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "product-id is not a valid UUID.",
		},
		{
			name: "missing quantity",
			request: CreateOrderRequest{
				ProductID: validUUID,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "quantity must be > 0.",
		},
		{
			name: "negative quantity",
			request: CreateOrderRequest{
				ProductID: validUUID,
				Quantity:  -100,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "quantity must be > 0.",
		},
		{
			name: "zero quantity",
			request: CreateOrderRequest{
				ProductID: validUUID,
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
	const validUUID = "f47ac10b-58cc-4372-a567-0e02b2c3d009"

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
				ProductID: validUUID,
				Quantity:  5,
				Status:    order.PAID,
				Version:   1,
			},
			expectError: false,
		},
		{
			name: "missing ProductID",
			request: UpdateOrderRequest{
				Quantity: 5,
				Status:   order.PAID,
				Version:  1,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "product-id cannot be blank.",
		},
		{
			name: "empty ProductID",
			request: UpdateOrderRequest{
				ProductID: "",
				Quantity:  5,
				Status:    order.PAID,
				Version:   1,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "product-id cannot be blank.",
		},
		{
			name: "blank ProductID",
			request: UpdateOrderRequest{
				ProductID: "   ",
				Quantity:  5,
				Status:    order.PAID,
				Version:   1,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "product-id cannot be blank.",
		},
		{
			name: "ProductID not valid UUID",
			request: UpdateOrderRequest{
				ProductID: "1",
				Quantity:  5,
				Status:    order.PAID,
				Version:   1,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "product-id is not a valid UUID.",
		},
		{
			name: "missing quantity",
			request: UpdateOrderRequest{
				ProductID: validUUID,
				Status:    order.PAID,
				Version:   1,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "quantity must be > 0.",
		},
		{
			name: "negative quantity",
			request: UpdateOrderRequest{
				ProductID: validUUID,
				Quantity:  -5,
				Status:    order.PAID,
				Version:   1,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "quantity must be > 0.",
		},
		{
			name: "zero quantity",
			request: UpdateOrderRequest{
				ProductID: validUUID,
				Quantity:  0,
				Status:    order.PAID,
				Version:   1,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "quantity must be > 0.",
		},
		{
			name: "missing status",
			request: UpdateOrderRequest{
				ProductID: validUUID,
				Quantity:  1,
				Version:   1,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "status is not valid.",
		},
		{
			name: "empty status",
			request: UpdateOrderRequest{
				ProductID: validUUID,
				Quantity:  1,
				Status:    order.OrderStatus(" "),
				Version:   1,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "status is not valid.",
		},
		{
			name: "invalid status",
			request: UpdateOrderRequest{
				ProductID: validUUID,
				Quantity:  1,
				Status:    order.OrderStatus("PIAD"),
				Version:   1,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "status is not valid.",
		},
		{
			name: "invalid status with lowercase status",
			request: UpdateOrderRequest{
				ProductID: validUUID,
				Quantity:  5,
				Status:    order.OrderStatus("paid"),
				Version:   1,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "status is not valid.",
		},
		{
			name: "missing version",
			request: UpdateOrderRequest{
				ProductID: validUUID,
				Quantity:  1,
				Status:    order.PAID,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "version must be > 0.",
		},
		{
			name: "negative version",
			request: UpdateOrderRequest{
				ProductID: validUUID,
				Quantity:  1,
				Status:    order.PAID,
				Version:   -5,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "version must be > 0.",
		},
		{
			name: "zero quantity",
			request: UpdateOrderRequest{
				ProductID: validUUID,
				Quantity:  1,
				Status:    order.PAID,
				Version:   0,
			},
			expectError:          true,
			numberOfErrors:       1,
			expectedErrorMessage: "version must be > 0.",
		},
		{
			name: "empty name, zero price, invalid status, zero quantity",
			request: UpdateOrderRequest{
				ProductID: "",
				Quantity:  -2,
				Status:    order.OrderStatus("PIAD"),
			},
			expectError:          true,
			numberOfErrors:       4,
			expectedErrorMessage: "product-id cannot be blank.; quantity must be > 0.; status is not valid.; version must be > 0.",
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
