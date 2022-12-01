package main

import (
	"database/sql"
	"errors"
	"net/http"

	lru "github.com/hashicorp/golang-lru"
	"github.com/sujamess/k6-the-hard-way/pkgs/httpparam"
	"github.com/sujamess/k6-the-hard-way/pkgs/httpwriter"
	productmodel "github.com/sujamess/k6-the-hard-way/pkgs/model/product"
	"github.com/sujamess/k6-the-hard-way/pkgs/randomer"
)

type handler struct {
	lru *lru.Cache
}

func NewHandler() handler {
	cache, err := lru.New(1e3)
	if err != nil {
		panic(err)
	}
	return handler{lru: cache}
}

func (h handler) MockProducts(w http.ResponseWriter, r *http.Request) {
	prices := randomer.Floats(1, 1e6, 1e5)
	products := make([]productmodel.Product, len(prices))
	for i, price := range prices {
		products[i] = productmodel.Product{Price: price}
	}

	err := BulkAddProducts(r.Context(), products)
	if err != nil {
		httpwriter.Write(w, http.StatusInternalServerError, err)
		return
	}

	httpwriter.Write(w, http.StatusCreated, nil)
}

func (h handler) ListProductsByIDs(w http.ResponseWriter, r *http.Request) {
	ids, err := httpparam.QueryParams[uint64](r, "ids")
	if err != nil {
		httpwriter.Write(w, http.StatusBadRequest, err)
		return
	}

	products, err := ListProductByIDs(r.Context(), ids)
	if errors.Is(err, sql.ErrNoRows) {
		httpwriter.Write(w, http.StatusNotFound, errors.New("product not found"))
		return
	} else if err != nil {
		httpwriter.Write(w, http.StatusInternalServerError, err)
		return
	}
	httpwriter.Write(w, http.StatusOK, products)
}

func (h handler) ListProductsByIDsWithCache(w http.ResponseWriter, r *http.Request) {
	ids, err := httpparam.QueryParams[uint64](r, "ids")
	if err != nil {
		httpwriter.Write(w, http.StatusBadRequest, err)
		return
	}

	res := make([]productmodel.Product, 0)
	notInCacheIDs := make([]uint64, 0)
	for _, id := range ids {
		if v, ok := h.lru.Get(id); ok {
			res = append(res, v.(productmodel.Product))
		} else {
			notInCacheIDs = append(notInCacheIDs, id)
		}
	}

	products, err := ListProductByIDs(r.Context(), notInCacheIDs)
	if errors.Is(err, sql.ErrNoRows) {
		httpwriter.Write(w, http.StatusNotFound, errors.New("product not found"))
		return
	} else if err != nil {
		httpwriter.Write(w, http.StatusInternalServerError, err)
		return
	}

	for _, p := range products {
		_ = h.lru.Add(p.ID, p)
	}

	res = append(res, products...)
	httpwriter.Write(w, http.StatusOK, res)
}

func (h handler) Healthcheck(w http.ResponseWriter, r *http.Request) {
	if err := Ping(r.Context()); err != nil {
		httpwriter.Write(w, http.StatusInternalServerError, err)
		return
	}
	httpwriter.Write(w, http.StatusNoContent, nil)
}
