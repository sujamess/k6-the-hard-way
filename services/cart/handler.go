package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/sujamess/k6-the-hard-way/pkgs/binder"
	"github.com/sujamess/k6-the-hard-way/pkgs/httpparam"
	"github.com/sujamess/k6-the-hard-way/pkgs/httpwriter"
	cartmodel "github.com/sujamess/k6-the-hard-way/pkgs/model/cart"
	ordermodel "github.com/sujamess/k6-the-hard-way/pkgs/model/order"
	"github.com/sujamess/k6-the-hard-way/pkgs/uniquer"
)

type handler struct {
	orderService OrderService
}

func NewHandler(orderService OrderService) handler {
	return handler{orderService: orderService}
}

func (h handler) CreateCart(w http.ResponseWriter, r *http.Request) {
	cartUUID := uniquer.UUID()
	err := CreateCart(r.Context(), cartUUID)
	if err != nil {
		httpwriter.Write(w, http.StatusInternalServerError, err)
		return
	}
	httpwriter.Write(w, http.StatusCreated, cartmodel.CreateCartResponse{CartUUID: cartUUID})
}

func (h handler) ListProductsInCart(w http.ResponseWriter, r *http.Request) {
	cartUUID, err := httpparam.PathParam[string](r, "uuid")
	if err != nil {
		httpwriter.Write(w, http.StatusBadRequest, err)
		return
	}

	pic, err := ListProductsInCart(r.Context(), cartmodel.ListProductInCartFilter{CartUUID: cartUUID})
	if errors.Is(err, sql.ErrNoRows) {
		httpwriter.Write(w, http.StatusNotFound, fmt.Errorf("products in cart id '%s' not found", cartUUID))
		return
	} else if err != nil {
		httpwriter.Write(w, http.StatusInternalServerError, err)
		return
	}

	httpwriter.Write(w, http.StatusOK, pic)
}

func (h handler) AddProductToCart(w http.ResponseWriter, r *http.Request) {
	cartUUID, err := httpparam.PathParam[string](r, "uuid")
	if err != nil {
		httpwriter.Write(w, http.StatusBadRequest, err)
		return
	}

	var req cartmodel.AddProductToCartRequest
	if err := binder.Bind[*cartmodel.AddProductToCartRequest](r, &req); err != nil {
		httpwriter.Write(w, http.StatusBadRequest, err)
		return
	}

	err = AddProductToCart(
		r.Context(),
		cartmodel.AddProductToCartRequest{
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
		},
		cartUUID,
	)
	if err != nil {
		httpwriter.Write(w, http.StatusInternalServerError, err)
		return
	}

	httpwriter.Write(w, http.StatusCreated, nil)
}

func (h handler) Checkout(w http.ResponseWriter, r *http.Request) {
	cartUUID, err := httpparam.PathParam[string](r, "uuid")
	if err != nil {
		httpwriter.Write(w, http.StatusBadRequest, err)
		return
	}

	pic, err := ListProductsInCart(r.Context(), cartmodel.ListProductInCartFilter{CartUUID: cartUUID})
	if errors.Is(err, sql.ErrNoRows) {
		httpwriter.Write(w, http.StatusNotFound, fmt.Errorf("products in cart id '%s' not found or already checked out", cartUUID))
		return
	} else if err != nil {
		httpwriter.Write(w, http.StatusInternalServerError, err)
		return
	}

	products := make([]ordermodel.Product, 0, len(pic))
	for i := range pic {
		products = append(products, ordermodel.Product{
			ProductID: &pic[i].ProductID,
			Quantity:  pic[i].Quantity,
		})
	}

	err = h.orderService.Create(ordermodel.CreateOrderRequest{CartUUID: cartUUID, Products: products})
	if err != nil {
		httpwriter.Write(w, http.StatusInternalServerError, err)
		return
	}

	httpwriter.Write(w, http.StatusCreated, nil)
}

func (h handler) CheckoutWithAsync(w http.ResponseWriter, r *http.Request) {
	cartUUID, err := httpparam.PathParam[string](r, "uuid")
	if err != nil {
		httpwriter.Write(w, http.StatusBadRequest, err)
		return
	}

	pic, err := ListProductsInCart(r.Context(), cartmodel.ListProductInCartFilter{CartUUID: cartUUID})
	if errors.Is(err, sql.ErrNoRows) {
		httpwriter.Write(w, http.StatusNotFound, fmt.Sprintf("products in cart id '%s' not found", cartUUID))
		return
	} else if err != nil {
		httpwriter.Write(w, http.StatusInternalServerError, err)
		return
	}

	products := make([]ordermodel.Product, 0, len(pic))
	for i := range pic {
		products = append(products, ordermodel.Product{
			ProductID: &pic[i].ProductID,
			Quantity:  pic[i].Quantity,
		})
	}

	err = UpdateCart(r.Context(), cartmodel.UpdateCartRequest{Status: "ORDER_PROCESSING"}, cartmodel.UpdateCartFilter{CartUUID: cartUUID})
	if err != nil {
		httpwriter.Write(w, http.StatusInternalServerError, err)
		return
	}

	err = h.orderService.CreateWithAsynchronous(ordermodel.CreateOrderRequest{CartUUID: cartUUID, Products: products})
	if err != nil {
		httpwriter.Write(w, http.StatusInternalServerError, err)
		return
	}
	httpwriter.Write(w, http.StatusCreated, nil)
}

func (h handler) UpdateCart(w http.ResponseWriter, r *http.Request) {
	cartUUID, err := httpparam.PathParam[string](r, "uuid")
	if err != nil {
		httpwriter.Write(w, http.StatusBadRequest, err)
		return
	}

	var req cartmodel.UpdateCartRequest
	if err := binder.Bind[*cartmodel.UpdateCartRequest](r, &req); err != nil {
		httpwriter.Write(w, http.StatusBadRequest, err)
		return
	}

	err = UpdateCart(r.Context(), req, cartmodel.UpdateCartFilter{CartUUID: cartUUID})
	if err != nil {
		httpwriter.Write(w, http.StatusInternalServerError, err)
		return
	}

	httpwriter.Write(w, http.StatusNoContent, nil)
}
