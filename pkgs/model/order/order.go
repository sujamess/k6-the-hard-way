package order

// Order
type Order struct {
	ID          uint64  `json:"orderID"`
	OrderNumber string  `json:"orderNumber"`
	CartUUID    string  `json:"cartUUID"`
	Amount      float64 `json:"amount"`
}

// OrderProduct
type OrderProduct struct {
	OrderID   uint64  `json:"orderID"`
	ProductID uint64  `json:"productID"`
	Quantity  uint64  `json:"quantity"`
	Amount    float64 `json:"amount"`
}
