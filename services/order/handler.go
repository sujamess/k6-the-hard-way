package main

import (
	"context"
	"net/http"

	"github.com/sujamess/k6-the-hard-way/pkgs/binder"
	"github.com/sujamess/k6-the-hard-way/pkgs/httpwriter"
	cartmodel "github.com/sujamess/k6-the-hard-way/pkgs/model/cart"
	ordermodel "github.com/sujamess/k6-the-hard-way/pkgs/model/order"
	"github.com/sujamess/k6-the-hard-way/pkgs/model/product"
	"github.com/sujamess/k6-the-hard-way/pkgs/uniquer"
)

type handler struct {
	cartService    CartService
	productService ProductService
}

func NewHandler(cartService CartService, productService ProductService) handler {
	return handler{cartService: cartService, productService: productService}
}

func (h handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req ordermodel.CreateOrderRequest
	if err := binder.Bind[*ordermodel.CreateOrderRequest](r, &req); err != nil {
		httpwriter.Write(w, http.StatusBadRequest, err)
		return
	}

	productIDs := make([]uint64, len(req.Products))
	for i, p := range req.Products {
		productIDs[i] = *p.ProductID
	}

	products, err := h.productService.ListProductsByIDs(r.Context(), productIDs)
	if err != nil {
		httpwriter.Write(w, http.StatusInternalServerError, err)
		return
	}

	err = h.createOrder(r.Context(), req, products)
	if err != nil {
		httpwriter.Write(w, http.StatusInternalServerError, err)
		return
	}
	httpwriter.Write(w, http.StatusCreated, nil)
}

func (h handler) createOrder(ctx context.Context, req ordermodel.CreateOrderRequest, products []product.Product) error {
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
	for i, p := range products {
		orderProducts[i] = ordermodel.OrderProduct{
			OrderID:   orderID,
			ProductID: p.ID,
			Quantity:  hash[p.ID],
			Amount:    float64(hash[p.ID]) * p.Price,
		}
	}
	err = BulkCreateOrderProduct(ctx, orderProducts)
	if err != nil {
		return err
	}

	return h.cartService.UpdateCartWithAsynchronous(cartmodel.UpdateCartFromBrokerRequest{
		Update: cartmodel.UpdateCartRequest{
			Status:      "ORDER_CREATED",
			OrderNumber: &orderNumber,
		},
		Filter: cartmodel.UpdateCartFilter{
			CartUUID: req.CartUUID,
		},
	})
}
