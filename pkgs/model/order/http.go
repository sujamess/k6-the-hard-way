package order

import (
	"fmt"

	"github.com/sujamess/k6-the-hard-way/pkgs/binder"
)

// Create Order
type (
	CreateOrderRequest struct {
		CartUUID string    `json:"cartUUID"`
		Products []Product `json:"products"`
	}
	Product struct {
		ProductID *uint64 `json:"productID"`
		Quantity  uint64  `json:"quantity"`
	}
)

func (req *CreateOrderRequest) Validate() error {
	if req.CartUUID == "" {
		return fmt.Errorf("cartUUID: %w", binder.ErrRequiredField)
	}
	if len(req.Products) == 0 {
		return fmt.Errorf("products: %w", binder.ErrRequiredField)
	}

	for _, p := range req.Products {
		if p.ProductID == nil {
			return fmt.Errorf("productID: %w", binder.ErrRequiredField)
		} else if p.Quantity == 0 {
			return fmt.Errorf("quantity: %w", binder.ErrRequiredField)
		}
	}

	return nil
}
