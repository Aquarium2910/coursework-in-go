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

			err = printDb(writer, controller, limit)
			if err != nil {
				slog.Error("couldn't print db",
					"error", err,
					"operation", "printDb",
					"limit", limit)
				return err
			}

		case "2":
			err = formNewOrder(writer, reader, controller)
			HandleError(writer, err)
			if err == nil {
				fmt.Fprintf(writer, "\nSuccessfully added new order!\n")
			}
		case "3":
			err = updateOrderType(writer, reader, controller)
			HandleError(writer, err)
			if err == nil {
				fmt.Fprintf(writer, "\nOrder updated successfully\n")
			}
		case "4":
			err = deleteOrder(writer, reader, controller)
			HandleError(writer, err)
			if err == nil {
				fmt.Fprintf(writer, "Order deleted successfully\n")
			}
		case "5":
			err = ShowDatesWithBiggestOrders(writer, controller, biggestOrdersLimit)
			HandleError(writer, err)
		case "6":
			err = ShowOrdersWhenRateChanged(writer, controller)
			HandleError(writer, err)
		case "7":
			err = ShowAvgNumOfOrdersLessThen(writer, controller, typeOfOrdersLessThan, lessThanThreshold)
			HandleError(writer, err)
		case "8":
			err = ShowTypesOfSmallestOrders(writer, controller, typesOfSmallestOrdersLimit)
			HandleError(writer, err)
		case "9":
			err = ShowStatsForPeriods(writer, controller)
			HandleError(writer, err)
		case "10":
			return fmt.Errorf("exit program")
		default:
			fmt.Fprintf(writer, frontend.InvalidChoice+"\n\n")
		}

	}
}

func HandleError(writer io.Writer, err error) {
	if err != nil {
		slog.Error("", "error", err)
		fmt.Fprintln(writer, "Something went wrong, try again.")
	}
}

func printDb(writer io.Writer, controller *postgres.DbController, limit int) error {
	orders, err := controller.SelectAllOrders(limit)
	if err != nil {
		return fmt.Errorf("couldn`t print db: %w", err)
	}

	frontend.PrintTable(writer, orders)
	return nil
}

func formNewOrder(writer io.Writer, reader io.Reader, controller *postgres.DbController) error {
	var order models.Order
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

	err = controller.AddNewOrder(order.TimeStamp, order.Type, order.Amount, order.Currency, order.ExchangeRate)

	if err != nil {
		return fmt.Errorf("failed to add new order: %w", err)
	}

	return nil
}

func updateOrderType(writer io.Writer, reader io.Reader, controller *postgres.DbController) error {
	var orderTypeNew string
	var orderId int

	input, err := frontend.TakeInput(writer, reader, "Enter order id: ")
	if err != nil {
		return fmt.Errorf("error updating order's type: %w", err)
	}
	orderId, err = strconv.Atoi(input)
	if err != nil {
		return fmt.Errorf("error updating order's type: %w", err)
	}

	input, err = frontend.TakeInput(writer, reader, "Enter new order type: ")
	if err != nil {
		return fmt.Errorf("error updating order's type: %w", err)
	}

	err = controller.UpdateOrder(orderId, orderTypeNew)
	if err != nil {
		return fmt.Errorf("error updating order's type: %w", err)
	}

	return nil
}

func deleteOrder(writer io.Writer, reader io.Reader, controller *postgres.DbController) error {
	var orderId int

	inputId, err := frontend.TakeInput(writer, reader, "Enter order id: ")
	if err != nil {
		return fmt.Errorf("error deleting order: %w", err)
	}

	orderId, err = strconv.Atoi(inputId)
	if err != nil {
		return fmt.Errorf("error deleting order: %w", err)
	}

	err = controller.DeleteOrder(orderId)
	if err != nil {
		return fmt.Errorf("error deleting order: %w", err)
	}

	return nil
}

func ShowDatesWithBiggestOrders(writer io.Writer, controller *postgres.DbController, limit int) error {
	orders, err := controller.DatesWithBiggestOrders(limit)
	if err != nil {
		return fmt.Errorf("couldn't show biggest orders: %w", err)
	}

	frontend.PrintBiggestOrders(writer, orders)

	return nil
}

func ShowTypesOfSmallestOrders(writer io.Writer, controller *postgres.DbController, limit int) error {
	orderTypes, err := controller.TypeOfSmallestOrders(limit)
	if err != nil {
		return fmt.Errorf("couldn't show types of smallest orders: %w", err)
	}

	fmt.Println("")
	for _, orderType := range orderTypes {
		fmt.Fprintf(writer, "%s\n", orderType)
	}

	return nil
}

func ShowOrdersWhenRateChanged(writer io.Writer, controller *postgres.DbController) error {
	rows, err := controller.OrdersWhenRateChanged()
	if err != nil {
		return fmt.Errorf("couldn't show orders when rate changed: %w", err)
	}

	frontend.PrintTable(writer, rows)

	return nil
}

func ShowAvgNumOfOrdersLessThen(writer io.Writer, controller *postgres.DbController, orderType string, lessThen float64) error {

	avgNum, err := controller.GetAvgNumOfOrdersLessThan(orderType, lessThen)
	if err != nil {
		return fmt.Errorf("couldn't show avg-num of orders: %w", err)
	}

	fmt.Fprintf(writer, "\nAvg num of orders of type %s per month less then %.2f: %.2f\n", orderType, lessThen, avgNum)

	return nil
}

func ShowStatsForPeriods(writer io.Writer, controller *postgres.DbController) error {

	stats, err := controller.GetTableForPeriods()
	if err != nil {
		fmt.Errorf("couldn't show stats: %w", err)
	}

	frontend.PrintStats(writer, stats)

	return nil
}
