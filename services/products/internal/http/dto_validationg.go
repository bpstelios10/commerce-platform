package handler

import (
	"fmt"
	"log/slog"
)

func validateCreateProduct(req CreateProductRequest) error {
	var errs []string

	if req.ID == "" {
		errs = append(errs, "id is required")
	}
	if req.Name == "" {
		errs = append(errs, "name is required")
	}
	if req.Price <= 0 {
		errs = append(errs, "price must be > 0")
	}

	if len(errs) > 0 {
		msg := fmt.Sprintf("%v", errs)

		slog.Warn("invalid create product request", "error", msg)

		return fmt.Errorf("%s", msg)
	}

	return nil
}

func validateUpdateProduct(req UpdateProductRequest) error {
	var errs []string

	if req.Name == "" {
		errs = append(errs, "name is required")
	}
	if req.Price <= 0 {
		errs = append(errs, "price must be > 0")
	}

	if len(errs) > 0 {
		msg := fmt.Sprintf("%v", errs)

		slog.Warn("invalid update product request", "error", msg)

		return fmt.Errorf("%s", msg)
	}

	return nil
}
