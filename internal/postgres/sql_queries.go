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

func (c *DbController) SelectAllOrders(limit int) ([]models.Order, error) {
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

	var orders []models.Order
	for rows.Next() {
		o := models.Order{}
		err = rows.Scan(&o.Id, &o.TimeStamp, &o.Type, &o.Amount, &o.Currency, &o.ExchangeRate)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("couldn't read order: %w", err)
	}

	return orders, nil
}

func (c *DbController) AddNewOrder(orderDate time.Time, orderType string, amount float64, currency string, exchangerate float64) error {
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

func (c *DbController) UpdateOrder(orderId int, orderType string) error {
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

func (c *DbController) DeleteOrder(orderId int) error {
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

func (c *DbController) DatesWithBiggestOrders(limit int) ([]models.BiggestOrders, error) {
	const query = `
		SELECT orderdate, SUM(amount*exchangerate) as total_uah FROM orders
		GROUP BY orderdate
		ORDER BY total_uah DESC 
		LIMIT $1`

	if limit == 0 {
		return nil, nil
	}

	rows, err := c.dbPool.Query(c.ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting dates with biggest orders: %w", err)
	}
	defer rows.Close()

	var orders []models.BiggestOrders
	for i := 0; rows.Next(); i++ {
		orders = append(orders, models.BiggestOrders{})
		err = rows.Scan(&orders[i].Date, &orders[i].TotalUah)
		if err != nil {
			return nil, fmt.Errorf("error getting dates with biggest orders: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error getting dates with biggest orders: %w", err)
	}
	return orders, nil
}

func (c *DbController) TypeOfSmallestOrders(limit int) ([]string, error) {
	const query = `SELECT DISTINCT ordertype FROM
                              (SELECT ordertype, amount FROM orders
                              ORDER BY amount ASC
                              LIMIT $1)`

	rows, err := c.dbPool.Query(c.ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting type of smallest orders: %w", err)
	}
	defer rows.Close()

	var orderTypes []string
	var res string
	for rows.Next() {
		err = rows.Scan(&res)
		if err != nil {
			return nil, fmt.Errorf("error getting type of smallest orders: %w", err)
		}
		orderTypes = append(orderTypes, res)
	}

	return orderTypes, nil
}

func (c *DbController) OrdersWhenRateChanged() ([]models.Order, error) {
	const query = `
		SELECT id, (orderdate + ordertime) as orderTimeStamp, ordertype, amount, currency, exchangerate 
		FROM orders
		WHERE (orderdate, currency) IN (
		SELECT orderdate, currency FROM orders
		GROUP BY orderdate, currency
		HAVING COUNT(DISTINCT exchangerate) > 1)
		ORDER BY orderdate, currency`

	rows, err := c.dbPool.Query(c.ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error getting orders when rate changed: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		o := models.Order{}
		err = rows.Scan(&o.Id, &o.TimeStamp, &o.Type, &o.Amount, &o.Currency, &o.ExchangeRate)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error getting orders when rate changed: %w", err)
	}

	return orders, nil
}

func (c *DbController) GetAvgNumOfOrdersLessThan(orderType string, lessThen float64) (float64, error) {
	const query = `SELECT (SELECT COUNT(*) FROM orders
				WHERE ordertype LIKE $1
				  AND (amount*exchangerate) < $2) * 1.0
		/
		(SELECT COUNT (DISTINCT to_char(orderdate, 'YYYY-MM'))
		FROM orders) AS  avg_less_then`

	row := c.dbPool.QueryRow(c.ctx, query, orderType, lessThen)

	var avgNum float64
	err := row.Scan(&avgNum)
	if err != nil {
		return 0, fmt.Errorf("couldn't show avg num of orders less then %f: %w", lessThen, err)
	}

	return avgNum, nil
}

func (c *DbController) GetTableForPeriods() ([]models.PeriodStats, error) {
	const query = `
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

	rows, err := c.dbPool.Query(c.ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error getting table for stats: %w", err)
	}
	defer rows.Close()

	var stats []models.PeriodStats
	for i := 0; rows.Next(); i++ {
		stats = append(stats, models.PeriodStats{})
		err = rows.Scan(&stats[i].TimePeriod, &stats[i].TotalSales, &stats[i].BigSales, &stats[i].SmallSales)
		if err != nil {
			return nil, fmt.Errorf("error getting table for stats: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error getting table for stats: %w", err)
	}

	return stats, nil
}
