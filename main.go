package main

import (
	"bufio"
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
	exitProgram   = "9. Exit program"

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
	for true {
		fmt.Fprintf(writer, listAllOrders+"\n")
		fmt.Fprintf(writer, addNewOrder+"\n")
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
		case "9":
			return fmt.Errorf("Exit program")
		default:
			fmt.Fprintf(writer, invalidChoice+"\n\n")
		}

	}
	return nil
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
	var orderTimeStampInput string
	var orderType string
	var amount float64
	var currency string
	var exchangeRate float64
	var orderTimeStamp time.Time

	scanner := bufio.NewScanner(reader)
	fmt.Fprintf(writer, "Write the date and time of order in following format: (%s)\n", timeFormat)
	fmt.Fprintf(writer, "Order date: ")
	if scanner.Scan() {
		orderTimeStampInput = scanner.Text()
	}
	var err = scanner.Err()
	if err != nil {
		return false
	}

	orderTimeStamp, err = time.Parse(timeFormat, orderTimeStampInput)
	if err != nil {
		return false
	}

	fmt.Fprintf(writer, "Order type: ")
	_, err = fmt.Fscan(reader, &orderType)
	if err != nil {
		return false
	}

	fmt.Fprintf(writer, "Pay amount: ")
	_, err = fmt.Fscan(reader, &amount)
	if err != nil {
		return false
	}

	fmt.Fprintf(writer, "Currency: ")
	_, err = fmt.Fscan(reader, &currency)
	if err != nil {
		return false
	}

	fmt.Fprintf(writer, "Exchange rate: ")
	_, err = fmt.Fscan(reader, &exchangeRate)
	if err != nil {
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
