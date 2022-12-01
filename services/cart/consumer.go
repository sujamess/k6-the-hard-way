package main

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/sujamess/k6-the-hard-way/pkgs/httpmiddleware"
	cartmodel "github.com/sujamess/k6-the-hard-way/pkgs/model/cart"
)

type consumer struct {
	mysql *sql.DB
}

func NewConsumer(mysql *sql.DB) consumer {
	return consumer{mysql: mysql}
}

func (c consumer) UpdateCart(payload []byte) error {
	ctx := c.newContext()
	var req cartmodel.UpdateCartFromBrokerRequest
	err := json.Unmarshal(payload, &req)
	if err != nil {
		return err
	}

	return UpdateCart(ctx, req.Update, req.Filter)
}

func (c consumer) newContext() context.Context {
	return context.WithValue(context.Background(), httpmiddleware.SQLCtxKey{}, c.mysql)
}
