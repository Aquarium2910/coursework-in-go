package main

import (
	"bufio"
	"coursework/internal/models"
	"coursework/internal/postgres"
	"fmt"
	"io"
	"os"
	"time"
)

const dbURL = "postgres://user:password@localhost:5432/mydb"

const (
	biggestOrdersLimit         = 5
	typesOfSmallestOrdersLimit = 6
)

const (
	listAllOrders         = "1. List all orders"
	addNewOrder           = "2. Add new order"
	updateOrder           = "3. Update order type"
	deleteOrderString     = "4. Delete order"
	biggestOrdersDates    = "5. Show 5 dates with biggest orders"
	ordersWhenRateChanged = "6. Show orders at days when exchangerate changed"
	typesOfSmallestOrders = "8. Show types of 6 smallest orders"
	exitProgram           = "9. Exit program"

	printLimit    = "How many items to print? (0 will print all items)"
	invalidChoice = "Invalid choice"

	timeFormat = "2006-01-02 15:04:05"
)

func main() {
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
		fmt.Fprintf(writer, "\n"+listAllOrders+"\n")
		fmt.Fprintf(writer, addNewOrder+"\n")
		fmt.Fprintf(writer, updateOrder+"\n")
		fmt.Fprintf(writer, deleteOrderString+"\n")
		fmt.Fprintf(writer, biggestOrdersDates+"\n")
		fmt.Fprintf(writer, ordersWhenRateChanged+"\n")
		fmt.Fprintf(writer, typesOfSmallestOrders+"\n")
		fmt.Fprintf(writer, exitProgram+"\n")

		var userChoice string
		fmt.Fscan(reader, &userChoice)

		switch userChoice {
		case "1":
			var limit int
			fmt.Fprintf(writer, printLimit+"\n")
			fmt.Fscan(reader, &limit)

			printDb(writer, controller, limit)

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
		case "8":
			ShowTypesOfSmallestOrders(writer, controller, typesOfSmallestOrdersLimit)
		case "9":
			return fmt.Errorf("Exit program")
		default:
			fmt.Fprintf(writer, invalidChoice+"\n\n")
		}

	}
}

func printDb(writer io.Writer, controller *postgres.DbController, limit int) {
	orders, err := controller.SelectAllOrders(limit)
	if err != nil {
		fmt.Fprintf(writer, "Couldn't select all orders: %v\n", err)
		return
	}
	defer orders.Close()

	fmt.Fprintf(writer, "\nId	Date and time of order	Order type	"+
		"Pay amount		Currency 	Exchange rate	\n")
	for orders.Next() {
		var order models.Order

		err = orders.Scan(&order.Id, &order.OrderTimeStamp, &order.OrderType, &order.Amount, &order.Currency, &order.ExchangeRate)

		fmt.Fprintf(writer, "%d %s %s %f %s %f\n", order.Id, order.OrderTimeStamp, order.OrderType, order.Amount,
			order.Currency, order.ExchangeRate)
	}
	fmt.Fprintf(writer, "\n")

	if err := orders.Err(); err != nil {
		fmt.Fprintf(writer, "Couldn't read order: %v\n", err)
	}
}

func formNewOrder(writer io.Writer, reader io.Reader, controller *postgres.DbController) bool {
	var order models.Order

	scanner := bufio.NewScanner(reader)
	fmt.Fprintf(writer, "Write the date and time of order in following format: (%s)\n", timeFormat)
	fmt.Fprintf(writer, "Order date: ")
	if scanner.Scan() {
		order.OrderTimeStampInput = scanner.Text()
	}
	var err = scanner.Err()
	if err != nil {
		return false
	}

	order.OrderTimeStamp, err = time.Parse(timeFormat, order.OrderTimeStampInput)
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
