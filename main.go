package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"coursework/internal/postgres"
)

type Order struct {
	id           int
	orderDate    string
	orderTime    string
	orderType    string
	amount       float64
	currency     string
	exchangeRate float64
}

const dbURL = "postgres://user:password@localhost:5432/mydb"

const (
	listAllOrders = "1. List all orders"
	addNewOrder   = "2. Add new order"

	printLimit    = "How many items to print? (0 will print all items)"
	invalidChoice = "Invalid choice"

	timeFormat = "2006-01-02 15:04:05"
)

func main() {
	controller := postgres.NewDbController(dbURL)

	fmt.Println("✅ Додаток підключено до БД")

	menu(os.Stdout, os.Stdin, controller)
}

func menu(writer io.Writer, reader io.Reader, controller *postgres.DbController) {
	for true {
		fmt.Fprintf(writer, listAllOrders+"\n")
		fmt.Fprintf(writer, addNewOrder+"\n")

		var userChoice string
		fmt.Fscan(reader, &userChoice)

		switch userChoice {
		case "1":
			var limit int
			fmt.Fprintf(writer, printLimit+"\n")
			fmt.Fscan(reader, &limit)

			printDb(writer, controller, limit)

		case "2":
			formNewOrder(writer, reader, controller)

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

	fmt.Fprintf(writer, "Id	Date and time of order	Order type	"+
		"Pay amount		Currency 	Exchange rate	\n")
	for orders.Next() {
		var id int
		var orderTimeStamp time.Time
		var orderType string
		var amount float64
		var currency string
		var exchangeRate float64

		err = orders.Scan(&id, &orderTimeStamp, &orderType, &amount, &currency, &exchangeRate)

		fmt.Fprintf(writer, "%d %s %s %f %s %f\n", id, orderTimeStamp, orderType, amount, currency, exchangeRate)
	}

	if err := orders.Err(); err != nil {
		fmt.Fprintf(writer, "Couldn't read order: %v\n", err)
	}
}

func formNewOrder(writer io.Writer, reader io.Reader, controller *postgres.DbController) bool {
	var orderTimeStamp time.Time
	var orderType string
	var amount float64
	var currency string
	var exchangeRate float64

	fmt.Fprintf(writer, "Write the date and time of order in following format: (%s)\n", timeFormat)
	fmt.Fprintf(writer, "Order date: ")
	_, err := fmt.Fscan(reader, &orderTimeStamp)
	if err != nil {
		return false
	}

	fmt.Fprintf(writer, "Order type: ")
	_, err1 := fmt.Fscan(reader, &orderType)
	if err1 != nil {
		return false
	}

	fmt.Fprintf(writer, "Pay amount: ")
	_, err2 := fmt.Fscan(reader, &amount)
	if err2 != nil {
		return false
	}

	fmt.Fprintf(writer, "Currency: ")
	_, err3 := fmt.Fscan(reader, &currency)
	if err3 != nil {
		return false
	}

	fmt.Fprintf(writer, "Exchange rate: ")
	_, err4 := fmt.Fscan(reader, &exchangeRate)
	if err4 != nil {
		return false
	}

	isAdded := controller.AddNewOrder(orderTimeStamp, orderType, amount, currency, exchangeRate)

	if isAdded {
		fmt.Fprintf(writer, "✅Added new order succesfully\n")
		return true
	}

	fmt.Fprintf(writer, "❌Failed to add new order\n")
	return false
}
