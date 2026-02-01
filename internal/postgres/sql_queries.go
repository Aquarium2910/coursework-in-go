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

	addNewOrder = `INSERT INTO orders (orderDate, orderTime, orderType, amount, currency, exchangerate)
		 VALUES (($1::timestamp)::date, ($1::timestamp)::time, $2, $3, $4, $5)`
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

func (c *DbController) DatesWithBiggestOrders(limit int) (pgx.Rows, error) {
	if limit == 0 {
		return nil, nil
	}

	return c.dbPool.Query(c.ctx, selectDatesWithBiggestOrders, limit)
}
