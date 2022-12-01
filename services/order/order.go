package main

import (
	"context"
	"strings"

	"github.com/sujamess/k6-the-hard-way/pkgs/db/mysql"
	ordermodel "github.com/sujamess/k6-the-hard-way/pkgs/model/order"
)

func CreateOrder(ctx context.Context, req ordermodel.Order) (uint64, error) {
	conn, tx, err := mysql.ConnOrTx(ctx)
	if err != nil {
		return 0, err
	}

	query := "INSERT INTO `order` (order_number, cart_uuid, amount) VALUES (?, UUID_TO_BIN(?), ?);"
	args := []any{req.OrderNumber, req.CartUUID, req.Amount}

	if tx != nil {
		res, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return 0, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return 0, err
		}
		return uint64(id), nil
	}

	defer conn.Close()
	res, err := conn.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return uint64(id), nil
}

func BulkCreateOrderProduct(ctx context.Context, req []ordermodel.OrderProduct) error {
	conn, tx, err := mysql.ConnOrTx(ctx)
	if err != nil {
		return err
	}

	query := "INSERT INTO order_product (order_id, product_id, quantity, amount) VALUES (?, ?, ?, ?)" + strings.Repeat(", (?, ?, ?, ?)", len(req)-1) + ";"
	args := make([]any, 0)
	for i := range req {
		args = append(args, req[i].OrderID, req[i].ProductID, req[i].Quantity, req[i].Amount)
	}

	if tx != nil {
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		err = tx.Commit()
		if err != nil {
			return err
		}
		return nil
	}

	defer conn.Close()
	_, err = conn.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	return nil
}
