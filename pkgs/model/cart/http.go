package cart

import (
	"fmt"

	"github.com/sujamess/k6-the-hard-way/pkgs/binder"
)

// Create cart
type CreateCartResponse struct {
	CartUUID string `json:"cartUUID"`
}

// List products in cart
type (
	ListProductInCartFilter struct {
		CartUUID string
	}
)

// Add product to cart
type AddProductToCartRequest struct {
	ProductID *uint64 `json:"productID"`
	Quantity  *uint64 `json:"quantity"`
}

func (req *AddProductToCartRequest) Validate() error {
	if req.ProductID == nil {
		return fmt.Errorf("productID: %w", binder.ErrRequiredField)
	} else if req.Quantity == nil || (req.Quantity != nil && *req.Quantity == 0) {
		return fmt.Errorf("quantity: %w", binder.ErrRequiredField)
	}

	return nil
}

// Update cart
type (
	UpdateCartRequest struct {
		Status      string  `json:"status"`
		OrderNumber *string `json:"orderNumber"`
	}
	UpdateCartFilter struct {
		CartUUID string
	}
)

func (req *UpdateCartRequest) Validate() error {
	if req.Status == "" {
		return fmt.Errorf("status: %w", binder.ErrRequiredField)
	}
	return nil
}
