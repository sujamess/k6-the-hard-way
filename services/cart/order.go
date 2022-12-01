package main

import (
	"encoding/json"

	"github.com/sujamess/k6-the-hard-way/pkgs/broker"
	"github.com/sujamess/k6-the-hard-way/pkgs/httprequester"
	ordermodel "github.com/sujamess/k6-the-hard-way/pkgs/model/order"
)

type OrderService interface {
	Create(req ordermodel.CreateOrderRequest) error
	CreateWithAsynchronous(req ordermodel.CreateOrderRequest) error
}

type orderService struct {
	requester httprequester.HTTPRequester
	producer  broker.Producer
}

func NewOrderService(requester httprequester.HTTPRequester, producer broker.Producer) OrderService {
	os := &orderService{requester: requester, producer: producer}
	return os
}

func (os *orderService) Create(req ordermodel.CreateOrderRequest) error {
	return os.requester.Post("/orders", req)
}

func (os *orderService) CreateWithAsynchronous(req ordermodel.CreateOrderRequest) error {
	bytes, err := json.Marshal(req)
	if err != nil {
		return err
	}
	os.producer.Publish(broker.CreateOrderTopic, bytes)
	return nil
}
