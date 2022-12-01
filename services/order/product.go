package main

import (
	"context"
	"strconv"
	"strings"

	"github.com/sujamess/k6-the-hard-way/pkgs/httprequester"
	productmodel "github.com/sujamess/k6-the-hard-way/pkgs/model/product"
)

type ProductService interface {
	ListProductsByIDs(ctx context.Context, ids []uint64) ([]productmodel.Product, error)
	ListProductsByIDsWithCache(ctx context.Context, ids []uint64) ([]productmodel.Product, error)
}

type productService struct {
	requester httprequester.HTTPRequester
}

func NewProductService(requester httprequester.HTTPRequester) ProductService {
	return &productService{requester: requester}
}

func (ps *productService) ListProductsByIDs(ctx context.Context, ids []uint64) ([]productmodel.Product, error) {
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = strconv.FormatUint(id, 10)
	}
	var res []productmodel.Product
	err := ps.requester.Get("/products?ids="+strings.Join(strs, ","), &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (ps *productService) ListProductsByIDsWithCache(ctx context.Context, ids []uint64) ([]productmodel.Product, error) {
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = strconv.FormatUint(id, 10)
	}
	var res []productmodel.Product
	err := ps.requester.Get("/products/with-cache?ids="+strings.Join(strs, ","), &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
