package main

import (
	"bufio"
	"coursework/internal/frontend"
	"coursework/internal/models"
	"coursework/internal/postgres"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	biggestOrdersLimit         = 5
	typesOfSmallestOrdersLimit = 6

	typeOfOrdersLessThan = "харчування"
	lessThanThreshold    = 50
)

func main() {
	cleanup, err := StartupLogger()
	if err != nil {
		panic(err)
	}
	defer cleanup()

	slog.Info("Logger initialized")

	_ = godotenv.Load()
	dbURL := os.Getenv("DB_URL")

	slog.Info("Loaded environment variables")

	controller := postgres.NewDbController(dbURL)
	defer controller.Close()

	slog.Info("✅ Connected to DB")

	err = menu(os.Stdout, os.Stdin, controller)

	if err != nil {
		log.Fatal(err)
	}
}

func StartupLogger() (func(), error) {
	logFile, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("couldn't create/open log file: %w", err)
	}

	cleanUp := func() {
		err = logFile.Close()
		if err != nil {
			return
		}
	}

	logger := slog.New(slog.NewTextHandler(logFile, nil))
	slog.SetDefault(logger)

	return cleanUp, nil
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

			err = printDbNew(writer, controller, limit)
			if err != nil {
				slog.Error("couldn't print db",
					"error", err,
					"operation", "printDbNew",
					"limit", limit)
				return err
			}

		case "2":
			err = formNewOrderNew(writer, reader, controller)
			if err != nil {
				slog.Error("error forming new order", "error", err)
				fmt.Fprintln(writer, "Something went wrong, try again.")
			} else {
				fmt.Fprintf(writer, "\nSuccessfully added new order!\n")
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
			ShowAvgNumOfOrdersLessThen(writer, controller, typeOfOrdersLessThan, lessThanThreshold)
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

func printDbNew(writer io.Writer, controller *postgres.DbController, limit int) error {
	orders, err := controller.SelectAllOrdersNew(limit)
	if err != nil {
		return fmt.Errorf("couldn`t print db: %w", err)
	}

	frontend.PrintTable(writer, orders)
	return nil
}

func formNewOrderNew(writer io.Writer, reader io.Reader, controller *postgres.DbController) error {
	var order models.Orders
	var OrderTimeStampInput string

	scanner := bufio.NewScanner(reader)
	fmt.Fprintf(writer, "Write the date and time of order in following format: (%s)\n", frontend.TimeFormat)
	fmt.Fprintf(writer, "Order date: ")
	if scanner.Scan() {
		OrderTimeStampInput = scanner.Text()
	}
	var err = scanner.Err()
	if err != nil {
		return fmt.Errorf("couldn't form new order: %w", err)
	}

	order.TimeStamp, err = time.Parse(frontend.TimeFormat, OrderTimeStampInput)
	if err != nil {
		return fmt.Errorf("couldn't form new order: %w", err)
	}

	fmt.Fprintf(writer, "Order type: ")
	_, err = fmt.Fscan(reader, &order.Type)
	if err != nil {
		return fmt.Errorf("couldn't form new order: %w", err)
	}

	fmt.Fprintf(writer, "Pay amount: ")
	_, err = fmt.Fscan(reader, &order.Amount)
	if err != nil {
		return fmt.Errorf("couldn't form new order: %w", err)
	}

	fmt.Fprintf(writer, "Currency: ")
	_, err = fmt.Fscan(reader, &order.Currency)
	if err != nil {
		return fmt.Errorf("couldn't form new order: %w", err)
	}

	fmt.Fprintf(writer, "Exchange rate: ")
	_, err = fmt.Fscan(reader, &order.ExchangeRate)
	if err != nil {
		return fmt.Errorf("couldn't form new order: %w", err)
	}

	err = controller.AddNewOrderNew(order.TimeStamp, order.Type, order.Amount, order.Currency, order.ExchangeRate)

	if err != nil {
		return fmt.Errorf("failed to add new order: %w", err)
	}

	return nil
}

func updateOrderType(writer io.Writer, reader io.Reader, controller *postgres.DbController) {
	var orderTypeNew string
	var orderId int

	input, err := frontend.TakeInput(writer, reader, "Enter order id: ")
	if err != nil {
		fmt.Fprintln(writer, err)
	}
	orderId, _ = strconv.Atoi(input)

	input, err = frontend.TakeInput(writer, reader, "Enter new order type: ")

	err = controller.UpdateOrder(orderId, orderTypeNew)
	if err != nil {
		fmt.Fprintf(writer, "Couldn't update order: %v\n", err)
		return
	}

	fmt.Fprintf(writer, "Order updated successfully\n\n")
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
