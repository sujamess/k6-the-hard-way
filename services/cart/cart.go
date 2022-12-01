package main

import (
	"context"
	"database/sql"

	"github.com/sujamess/k6-the-hard-way/pkgs/db/mysql"
	cartmodel "github.com/sujamess/k6-the-hard-way/pkgs/model/cart"
)

func CreateCart(ctx context.Context, uuid string) error {
	conn, err := mysql.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, "INSERT INTO cart (uuid) VALUES (UUID_TO_BIN(?));", uuid)
	if err != nil {
		return err
	}
	return nil
}

func UpdateCart(ctx context.Context, update cartmodel.UpdateCartRequest, filter cartmodel.UpdateCartFilter) error {
	conn, err := mysql.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, "UPDATE cart SET status = ?, order_number = ? WHERE uuid = UUID_TO_BIN(?);", update.Status, update.OrderNumber, filter.CartUUID)
	if err != nil {
		return err
	}
	return nil
}

func AddProductToCart(ctx context.Context, req cartmodel.AddProductToCartRequest, cartUUID string) error {
	conn, err := mysql.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, `
		INSERT INTO cart_product (cart_uuid, product_id, quantity)
		VALUES (UUID_TO_BIN(?), ?, ?)
		AS new ON DUPLICATE KEY UPDATE quantity = new.quantity;
	`, cartUUID, *req.ProductID, *req.Quantity)
	if err != nil {
		return err
	}
	return nil
}

func ListProductsInCart(ctx context.Context, filter cartmodel.ListProductInCartFilter) ([]cartmodel.CartProduct, error) {
	conn, err := mysql.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.QueryContext(ctx, `
		SELECT cart_uuid, product_id, quantity
		FROM cart_product cp
		INNER JOIN cart c
		ON c.uuid = cp.cart_uuid
		WHERE (
			(c.status IS NULL OR c.status NOT IN ('ORDER_PROCESSING', 'ORDER_CREATED'))
			AND c.uuid = UUID_TO_BIN(?)
		);
	`, filter.CartUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cp := make([]cartmodel.CartProduct, 0)
	for rows.Next() {
		var res cartmodel.CartProduct
		err = rows.Scan(&res.CartUUID, &res.ProductID, &res.Quantity)
		if err != nil {
			return nil, err
		}
		cp = append(cp, res)
	}

	if len(cp) == 0 {
		return nil, sql.ErrNoRows
	}
	return cp, nil
}
