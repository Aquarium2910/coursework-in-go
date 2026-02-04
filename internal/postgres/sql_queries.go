package postgres

import (
	"context"
	"coursework/internal/models"
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
	selectDatesWithBiggestOrders = `
		SELECT orderdate, SUM(amount*exchangerate) as total_uah FROM orders
		GROUP BY orderdate
		ORDER BY total_uah DESC 
		LIMIT $1`
	selectTypeOfSmallestOrders = `SELECT DISTINCT ordertype FROM
                              (SELECT ordertype, amount FROM orders
                              ORDER BY amount ASC
                              LIMIT $1)`
	selectOrdersWhenRateChanges = `SELECT id, (orderdate + ordertime) as orderTimeStamp, ordertype, amount, currency, exchangerate 
	FROM orders
	WHERE (orderdate, currency) IN (
    SELECT orderdate, currency FROM ORDERS
    GROUP BY orderdate, currency
    HAVING COUNT(DISTINCT exchangerate) > 1)
	ORDER BY orderdate, currency`
	avgNOfOrdersFoodLessThan = `SELECT (SELECT COUNT(*) FROM orders
        WHERE ordertype LIKE $1
          AND (amount*exchangerate) < $2) * 1.0
/
(SELECT COUNT (DISTINCT to_char(orderdate, 'YYYY-MM'))
FROM orders) AS  avg_less_then`

	//Returns table with statistics for each period
	statsFor8HrPeriods = `
	SELECT
		CASE
			WHEN ordertime >= '00:00:00' AND ordertime < '08:00:00' THEN '00:00 - 08:00'
			WHEN ordertime >= '08:00:00' AND ordertime < '16:00:00' THEN '08:00 - 16:00'
			ELSE '16:00 - 23:59'
		END AS time_period,
	
		COUNT(*) AS total_sales,
	
		COUNT(*) FILTER (WHERE(amount*exchangerate) > 1000) AS big_sales,
	
		COUNT(*) FILTER (WHERE(amount*exchangerate) <= 1000) AS small_sales
	FROM orders
	GROUP BY
		1
	ORDER BY time_period`
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

func (c *DbController) SelectAllOrdersNew(limit int) ([]models.Orders, error) {
	const (
		query          = "SELECT id, (orderdate + ordertime) as orderTimeStamp, ordertype, amount, currency, exchangerate FROM orders"
		queryWithLimit = `SELECT id, (orderdate + ordertime) as orderTimeStamp, ordertype, amount, currency, exchangerate 
		 FROM orders
		 LIMIT $1`
	)

	var rows pgx.Rows
	var err error

	if limit == 0 {
		rows, err = c.dbPool.Query(c.ctx, query)
	} else {
		rows, err = c.dbPool.Query(c.ctx, queryWithLimit, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("error selecting all orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Orders
	for i := 0; rows.Next(); i++ {
		orders = append(orders, models.Orders{})
		err = rows.Scan(&orders[i].Id, &orders[i].TimeStamp, &orders[i].Type, &orders[i].Amount, &orders[i].Currency,
			&orders[i].ExchangeRate)

		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("couldn't read order: %w", err)
	}

	return orders, nil
}

func (c *DbController) AddNewOrderNew(orderDate time.Time, orderType string, amount float64, currency string, exchangerate float64) error {
	const query = `INSERT INTO orders (orderDate, orderTime, orderType, amount, currency, exchangerate)
		 VALUES (($1::timestamp)::date, ($1::timestamp)::time, $2, $3, $4, $5)`

	tag, err := c.dbPool.Exec(c.ctx, query, orderDate, orderType, amount, currency, exchangerate)

	if err != nil {
		return fmt.Errorf("error adding new order: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("error adding new order: no rows inserted")
	}

	return nil
}

func (c *DbController) UpdateOrderNew(orderId int, orderType string) error {
	const query = `UPDATE orders 
					SET ordertype = $1
					WHERE id = $2`

	report, err := c.dbPool.Exec(c.ctx, query, orderType, orderId)
	if err != nil {
		return fmt.Errorf("error updating order: %w", err)
	}

	if report.RowsAffected() == 0 {
		return fmt.Errorf("error updating order: row with id %d is not found", orderId)
	}

	return nil
}

func (c *DbController) DeleteOrderNew(orderId int) error {
	const query = `DELETE FROM orders WHERE id = $1`

	report, err := c.dbPool.Exec(c.ctx, query, orderId)
	if err != nil {
		return fmt.Errorf("error deleting order: %w", err)
	}

	if report.RowsAffected() == 0 {
		return fmt.Errorf("error deleting order: row with id %d is not found", orderId)
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

func (c *DbController) OrdersWhenRateChanged() (pgx.Rows, error) {
	rows, err := c.dbPool.Query(c.ctx, selectOrdersWhenRateChanges)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (c *DbController) GetAvgNumOfOrdersLessThan(orderType string, lessThen float64) pgx.Row {
	row := c.dbPool.QueryRow(c.ctx, avgNOfOrdersFoodLessThan, orderType, lessThen)
	return row
}

func (c *DbController) GetTableForPeriods() (pgx.Rows, error) {
	rows, err := c.dbPool.Query(c.ctx, statsFor8HrPeriods)
	if err != nil {
		return nil, err
	}

	return rows, nil
}
