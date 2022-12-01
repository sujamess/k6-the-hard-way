package main

import (
	"context"
	"database/sql"
	"strings"

	"github.com/sujamess/k6-the-hard-way/pkgs/db/mysql"
	"github.com/sujamess/k6-the-hard-way/pkgs/model/product"
)

func ListProductByIDs(ctx context.Context, ids []uint64) ([]product.Product, error) {
	conn, err := mysql.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	rows, err := conn.QueryContext(ctx, "SELECT id, price FROM product WHERE id IN (?"+strings.Repeat(`,?`, len(ids)-1)+")", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]product.Product, 0)
	for rows.Next() {
		var p product.Product
		err = rows.Scan(&p.ID, &p.Price)
		if err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	if len(res) == 0 {
		return nil, sql.ErrNoRows
	}
	return res, nil
}

func BulkAddProducts(ctx context.Context, products []product.Product) error {
	conn, err := mysql.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	prices := make([]any, len(products))
	for i, p := range products {
		prices[i] = p.Price
	}

	query := "INSERT INTO product (price) VALUES (?)" + strings.Repeat(", (?)", len(products)-1) + ";"

	_, err = conn.ExecContext(ctx, query, prices...)
	if err != nil {
		return err
	}
	return nil
}

func Ping(ctx context.Context) error {
	conn, err := mysql.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	return conn.PingContext(ctx)
}
