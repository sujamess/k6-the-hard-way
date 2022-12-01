package main

import (
	"encoding/json"

	"github.com/sujamess/k6-the-hard-way/pkgs/broker"
	"github.com/sujamess/k6-the-hard-way/pkgs/httprequester"
	cartmodel "github.com/sujamess/k6-the-hard-way/pkgs/model/cart"
)

type CartService interface {
	UpdateCart(cartUUID string, req cartmodel.UpdateCartRequest) error
	UpdateCartWithAsynchronous(req cartmodel.UpdateCartFromBrokerRequest) error
}

type cartService struct {
	requester httprequester.HTTPRequester
	producer  broker.Producer
}

func NewCartService(requester httprequester.HTTPRequester, producer broker.Producer) CartService {
	return cartService{requester: requester, producer: producer}
}

func (cs cartService) UpdateCart(cartUUID string, req cartmodel.UpdateCartRequest) error {
	return cs.requester.Patch("/carts/"+cartUUID, req)
}

func (cs cartService) UpdateCartWithAsynchronous(req cartmodel.UpdateCartFromBrokerRequest) error {
	bytes, err := json.Marshal(req)
	if err != nil {
		return err
	}
	cs.producer.Publish(broker.UpdateCartTopic, bytes)
	return nil
}
