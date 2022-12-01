package cart

// cart
type Cart struct {
	UUID        string  `json:"uuid"`
	Status      *string `json:"status"` // ORDER_PROCESSING, ORDER_CREATED
	OrderNumber *string `json:"orderNumber"`
}

// cart products
type CartProduct struct {
	CartUUID  string `json:"-"`
	ProductID uint64 `json:"productID"`
	Quantity  uint64 `json:"quantity"`
}
