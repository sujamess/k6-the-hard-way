package main

import (
	"database/sql"
	"encoding/json"

	"github.com/sujamess/k6-the-hard-way/pkgs/httpmiddleware"
	cartmodel "github.com/sujamess/k6-the-hard-way/pkgs/model/cart"
	ordermodel "github.com/sujamess/k6-the-hard-way/pkgs/model/order"
	"github.com/sujamess/k6-the-hard-way/pkgs/model/product"
	"github.com/sujamess/k6-the-hard-way/pkgs/uniquer"
	"golang.org/x/net/context"
)

type consumer struct {
	mysql          *sql.DB
	cartService    CartService
	productService ProductService
}

func NewConsumer(mysql *sql.DB, cartService CartService, productService ProductService) consumer {
	return consumer{mysql: mysql, cartService: cartService, productService: productService}
}

func (c consumer) CreateOrder(payload []byte) error {
	ctx, err := c.newContext()
	if err != nil {
		return err
	}

	var req ordermodel.CreateOrderRequest
	err = json.Unmarshal(payload, &req)
	if err != nil {
		return err
	}

	productIDs := make([]uint64, len(req.Products))
	for i, p := range req.Products {
		productIDs[i] = *p.ProductID
	}

	products, err := c.productService.ListProductsByIDs(ctx, productIDs)
	if err != nil {
		return err
	}

	err = c.createOrder(ctx, req, products)
	if err != nil {
		return err
	}

	return nil
}

func (c consumer) newContext() (context.Context, error) {
	ctx := context.Background()
	tx, err := c.mysql.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return context.WithValue(context.Background(), httpmiddleware.SQLTxCtxKey{}, tx), nil
}

func (c consumer) createOrder(ctx context.Context, req ordermodel.CreateOrderRequest, products []product.Product) error {
	hash := make(map[uint64]uint64, 0)
	for _, p := range req.Products {
		hash[*p.ProductID] = p.Quantity
	}

	var orderAmount float64
	for _, p := range products {
		orderAmount += float64(hash[p.ID]) * p.Price
	}

	orderNumber := uniquer.OrderNumber()
	orderID, err := CreateOrder(ctx, ordermodel.Order{
		OrderNumber: orderNumber,
		CartUUID:    req.CartUUID,
		Amount:      orderAmount,
	})
	if err != nil {
		return err
	}

	orderProducts := make([]ordermodel.OrderProduct, len(products))
	for i := range products {
		orderProducts[i] = ordermodel.OrderProduct{
			OrderID:   orderID,
			ProductID: products[i].ID,
			Quantity:  hash[products[i].ID],
			Amount:    float64(hash[products[i].ID]) * products[i].Price,
		}
	}
	err = BulkCreateOrderProduct(ctx, orderProducts)
	if err != nil {
		return err
	}

	return c.cartService.UpdateCartWithAsynchronous(cartmodel.UpdateCartFromBrokerRequest{
		Update: cartmodel.UpdateCartRequest{Status: "ORDER_CREATED", OrderNumber: &orderNumber},
		Filter: cartmodel.UpdateCartFilter{CartUUID: req.CartUUID},
	})
}
