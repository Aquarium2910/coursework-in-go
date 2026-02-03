package main

import (
	"bufio"
	"coursework/internal/frontend"
	"coursework/internal/models"
	"coursework/internal/postgres"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const (
	biggestOrdersLimit         = 5
	typesOfSmallestOrdersLimit = 6

	typeOfOrdersLessThan = "харчування"
	lessThanЕhreshold    = 50
)

func main() {
	_ = godotenv.Load()
	dbURL := os.Getenv("DB_URL")

	controller := postgres.NewDbController(dbURL)
	defer controller.Close()

	fmt.Println("✅ Додаток підключено до БД")

	err := menu(os.Stdout, os.Stdin, controller)

	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
}

func menu(writer io.Writer, reader io.Reader, controller *postgres.DbController) error {
	for {
		frontend.PrintOptions(writer)

		var userChoice string
		_, err := fmt.Fscan(reader, &userChoice)
		if err != nil {
			return err
		}

		switch userChoice {
		case "1":
			var limit int
			fmt.Fprintf(writer, frontend.PrintLimit+"\n")
			fmt.Fscan(reader, &limit)

			printDbNew(writer, controller, limit)

		case "2":
			isOk := formNewOrder(writer, reader, controller)
			if !isOk {
				fmt.Fprintf(writer, "Some error happened while forming order\n")
			}
		case "3":
			updateOrderType(writer, reader, controller)
		case "4":
			deleteOrder(writer, reader, controller)
		case "5":
			ShowDatesWithBiggestOrders(writer, controller, biggestOrdersLimit)
		case "6":
			ShowOrdersWhenRateChanged(writer, controller)
		case "7":
			ShowAvgNumOfOrdersLessThen(writer, controller, typeOfOrdersLessThan, lessThanЕhreshold)
		case "8":
			ShowTypesOfSmallestOrders(writer, controller, typesOfSmallestOrdersLimit)
		case "9":
			ShowStatsForPeriods(writer, controller)
		case "10":
			return fmt.Errorf("Exit program")
		default:
			fmt.Fprintf(writer, frontend.InvalidChoice+"\n\n")
		}

	}
}

func printDbNew(writer io.Writer, controller *postgres.DbController, limit int) {
	orders, err := controller.SelectAllOrdersNew(limit)
	if err != nil {
		fmt.Fprintln(writer, err)
	}

	frontend.PrintTable(writer, orders)
}

func formNewOrder(writer io.Writer, reader io.Reader, controller *postgres.DbController) bool {
	var order models.Order

	scanner := bufio.NewScanner(reader)
	fmt.Fprintf(writer, "Write the date and time of order in following format: (%s)\n", frontend.TimeFormat)
	fmt.Fprintf(writer, "Order date: ")
	if scanner.Scan() {
		order.OrderTimeStampInput = scanner.Text()
	}
	var err = scanner.Err()
	if err != nil {
		return false
	}

	order.OrderTimeStamp, err = time.Parse(frontend.TimeFormat, order.OrderTimeStampInput)
	if err != nil {
		return false
	}

	fmt.Fprintf(writer, "Order type: ")
	_, err = fmt.Fscan(reader, &order.OrderType)
	if err != nil {
		return false
	}

	fmt.Fprintf(writer, "Pay amount: ")
	_, err = fmt.Fscan(reader, &order.Amount)
	if err != nil {
		return false
	}

	fmt.Fprintf(writer, "Currency: ")
	_, err = fmt.Fscan(reader, &order.Currency)
	if err != nil {
		return false
	}

	fmt.Fprintf(writer, "Exchange rate: ")
	_, err = fmt.Fscan(reader, &order.ExchangeRate)
	if err != nil {
		return false
	}

	isAdded := controller.AddNewOrder(order.OrderTimeStamp, order.OrderType, order.Amount, order.Currency, order.ExchangeRate)

	if isAdded {
		fmt.Fprintf(writer, "✅Added new order succesfully\n")
		return true
	}

	fmt.Fprintf(writer, "❌Failed to add new order\n")
	return false
}

func ShowDatesWithBiggestOrders(writer io.Writer, controller *postgres.DbController, limit int) {
	rows, err := controller.DatesWithBiggestOrders(limit)
	if err != nil {
		fmt.Fprintf(writer, "Couldn't show biggest orders: %v\n", err)
	}

	defer rows.Close()

	var order models.Order

	fmt.Fprintf(writer, "Order dates		Total in UAH\n")
	for rows.Next() {
		err = rows.Scan(&order.OrderTimeStamp, &order.TotalUah)
		fmt.Fprintf(writer, "%s\t\t%f\n", order.OrderTimeStamp.Format(time.DateOnly), order.TotalUah)
	}
	fmt.Fprintf(writer, "\n")

	if err := rows.Err(); err != nil {
		fmt.Fprintf(writer, "Couldn't show biggest orders: %v\n", err)
	}
}

func updateOrderType(writer io.Writer, reader io.Reader, controller *postgres.DbController) {
	var orderTypeNew string
	var orderId int

	fmt.Fprintf(writer, "Enter order id: ")
	_, err := fmt.Fscan(reader, &orderId)
	if err != nil {
		fmt.Fprintf(writer, "Couldn't update order: %v\n", err)
		return
	}

	fmt.Fprintf(writer, "Enter new order type: ")

	_, err = fmt.Fscan(reader, &orderTypeNew)
	if err != nil {
		fmt.Fprintf(writer, "Couldn't update order: %v\n", err)
		return
	}

	err = controller.UpdateOrder(orderId, orderTypeNew)
	if err != nil {
		fmt.Fprintf(writer, "Couldn't update order: %v\n", err)
		return
	}

	fmt.Fprintf(writer, "Order updated successfully\n\n")
}

func deleteOrder(writer io.Writer, reader io.Reader, controller *postgres.DbController) {
	var orderId int
	fmt.Fprintf(writer, "\nEnter order id: ")

	_, err := fmt.Fscan(reader, &orderId)
	if err != nil {
		fmt.Fprintf(writer, "Couldn't delete order: %v\n", err)
	}

	err = controller.DeleteOrder(orderId)
	if err != nil {
		fmt.Fprintf(writer, "Couldn't delete order: %v\n", err)
	}

	fmt.Fprintf(writer, "Order deleted successfully\n\n")
}

func ShowTypesOfSmallestOrders(writer io.Writer, controller *postgres.DbController, limit int) {
	var orderType string
	rows, err := controller.TypeOfSmallestOrders(limit)

	if err != nil {
		fmt.Fprintf(writer, "Couldn't show types of smallest orders: %v\n", err)
		return
	}

	defer rows.Close()

	fmt.Fprintf(writer, "\n")
	for rows.Next() {
		err = rows.Scan(&orderType)
		if err != nil {
			fmt.Fprintf(writer, "Couldn't show types of smallest orders: %v\n", err)
		}

		fmt.Fprintf(writer, "%s\n", orderType)
	}
}

func ShowOrdersWhenRateChanged(writer io.Writer, controller *postgres.DbController) {
	var order models.Order
	rows, err := controller.OrdersWhenRateChanged()

	if err != nil {
		fmt.Fprintf(writer, "Couldn't show orders when rate changed: %v\n", err)
	}
	defer rows.Close()

	fmt.Fprintf(writer, "\nId	Date and time of order	Order type	"+
		"Pay amount		Currency 	Exchange rate	\n")

	for rows.Next() {
		err = rows.Scan(&order.Id, &order.OrderTimeStamp, &order.OrderType, &order.Amount, &order.Currency, &order.ExchangeRate)
		fmt.Fprintf(writer, "%d %s %s %f %s %f\n", order.Id, order.OrderTimeStamp, order.OrderType, order.Amount,
			order.Currency, order.ExchangeRate)
	}
	fmt.Fprintf(writer, "\n")

	if err = rows.Err(); err != nil {
		fmt.Fprintf(writer, "Couldn't read order: %v\n", err)
	}
}

func ShowAvgNumOfOrdersLessThen(writer io.Writer, controller *postgres.DbController, orderType string, lessThen float64) {
	var avgNum float64
	row := controller.GetAvgNumOfOrdersLessThan(orderType, lessThen)
	if row == nil {
		fmt.Fprintf(writer, "Couldn't show avg num of orders less then %f.\n", lessThen)
	}

	err := row.Scan(&avgNum)
	if err != nil {
		fmt.Fprintf(writer, "Error while scanning avg num: %v\n", err)
	}

	fmt.Fprintf(writer, "\nAvg num of orders of type %s per month less then %.2f: %.2f\n", orderType, lessThen, avgNum)
}

func ShowStatsForPeriods(writer io.Writer, controller *postgres.DbController) {
	var timePeriod string
	var totalSales int
	var bigSales int
	var smallSales int

	rows, err := controller.GetTableForPeriods()
	if err != nil {
		fmt.Fprintf(writer, "Couldn't show stats for periods: %v\n", err)
	}
	defer rows.Close()

	fmt.Fprintf(writer, "\nTime period\t\t Total\t   Big\t  Small\n")
	for rows.Next() {
		err = rows.Scan(&timePeriod, &totalSales, &bigSales, &smallSales)
		if err != nil {
			fmt.Fprintf(writer, "Couldn't show stats for periods: %v\n", err)
		}

		fmt.Fprintf(writer, "%s\t%5d\t%5d\t%5d\n", timePeriod, totalSales, bigSales, smallSales)
	}
}
