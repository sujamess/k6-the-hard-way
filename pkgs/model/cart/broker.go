package cart

type UpdateCartFromBrokerRequest struct {
	Update UpdateCartRequest `json:"update"`
	Filter UpdateCartFilter  `json:"filter"`
}
