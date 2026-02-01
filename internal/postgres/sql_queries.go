package postgres

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DbController struct {
	ctx    context.Context
	dbPool *pgxpool.Pool
}

const (
	selectAllOrders          = "SELECT id, (orderdate + ordertime) as orderTimeStamp, ordertype, amount, currency, exchangerate FROM orders"
	selectAllOrdersWithLimit = `SELECT id, (orderdate + ordertime) as orderTimeStamp, ordertype, amount, currency, exchangerate 
		 FROM orders
		 LIMIT $1`
	selectDatesWithBiggestOrders = `
		SELECT orderdate, SUM(amount*exchangerate) as total_uah FROM orders
		GROUP BY orderdate
		ORDER BY total_uah DESC 
		LIMIT $1`
	selectTypeOfSmallestOrders = `SELECT DISTINCT ordertype FROM
                              (SELECT ordertype, amount FROM orders
                              ORDER BY amount ASC
                              LIMIT $1)`

	addNewOrder = `INSERT INTO orders (orderDate, orderTime, orderType, amount, currency, exchangerate)
		 VALUES (($1::timestamp)::date, ($1::timestamp)::time, $2, $3, $4, $5)`

	updateOrder = `UPDATE orders 
					SET ordertype = $1
					WHERE id = $2`

	deleteOrder = `DELETE FROM orders WHERE id = $1`
)

func NewDbController(dbURL string) *DbController {
	ctx := context.Background()
	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}

	return &DbController{ctx: ctx, dbPool: dbPool}
}

func (c *DbController) Close() {
	c.dbPool.Close()
}

func (c *DbController) SelectAllOrders(limit int) (pgx.Rows, error) {
	if limit == 0 {
		return c.dbPool.Query(c.ctx, selectAllOrders)
	}

	return c.dbPool.Query(c.ctx, selectAllOrdersWithLimit, limit)
}

func (c *DbController) AddNewOrder(orderDate time.Time, orderType string, amount float64, currency string, exchangerate float64) bool {
	tag, err := c.dbPool.Exec(c.ctx, addNewOrder, orderDate, orderType, amount, currency, exchangerate)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error adding new order: %v\n", err)
		return false
	}

	if tag.RowsAffected() == 0 {
		fmt.Fprintf(os.Stderr, "Error adding new order: no rows inserted\n")
		return false
	}

	return true
}

func (c *DbController) UpdateOrder(orderId int, orderType string) error {
	report, err := c.dbPool.Exec(c.ctx, updateOrder, orderType, orderId)
	if err != nil {
		return err
	}

	if report.RowsAffected() == 0 {
		fmt.Fprintf(os.Stderr, "Error updating order: row with id %d is not found\n", orderId)
	}

	return nil
}

func (c *DbController) DeleteOrder(orderId int) error {
	report, err := c.dbPool.Exec(c.ctx, deleteOrder, orderId)
	if err != nil {
		return err
	}

	if report.RowsAffected() == 0 {
		fmt.Fprintf(os.Stderr, "Error deleting order: row with id %d is not found\n", orderId)
	}

	return nil
}

func (c *DbController) DatesWithBiggestOrders(limit int) (pgx.Rows, error) {
	if limit == 0 {
		return nil, nil
	}

	return c.dbPool.Query(c.ctx, selectDatesWithBiggestOrders, limit)
}

func (c *DbController) TypeOfSmallestOrders(limit int) (pgx.Rows, error) {
	rows, err := c.dbPool.Query(c.ctx, selectTypeOfSmallestOrders, limit)
	if err != nil {
		return nil, err
	}

	return rows, nil
}
