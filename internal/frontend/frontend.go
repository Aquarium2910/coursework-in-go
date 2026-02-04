package frontend

import (
	"coursework/internal/models"
	"fmt"
	"io"
)

const (
	listAllOrders          = "1. List all orders"
	addNewOrder            = "2. Add new order"
	updateOrder            = "3. Update order type"
	deleteOrderString      = "4. Delete order"
	biggestOrdersDates     = "5. Show 5 dates with biggest orders"
	ordersWhenRateChanged  = "6. Show orders at days when exchange rate changed"
	avgNumOrdersLessThen50 = "7. Show avg number of orders of type food less than 50 UAH per months"
	typesOfSmallestOrders  = "8. Show types of 6 smallest orders"
	statsFor8HrPeriods     = "9. Show stats for 8 hours periods"
	exitProgram            = "10. Exit program"

	PrintLimit    = "How many items to print? (0 will print all items)"
	InvalidChoice = "Invalid choice"

	TimeFormat = "2006-01-02 15:04:05"
)

func PrintOptions(writer io.Writer) {
	fmt.Fprintf(writer, "\n"+listAllOrders+"\n")
	fmt.Fprintf(writer, addNewOrder+"\n")
	fmt.Fprintf(writer, updateOrder+"\n")
	fmt.Fprintf(writer, deleteOrderString+"\n")
	fmt.Fprintf(writer, biggestOrdersDates+"\n")
	fmt.Fprintf(writer, ordersWhenRateChanged+"\n")
	fmt.Fprintf(writer, avgNumOrdersLessThen50+"\n")
	fmt.Fprintf(writer, typesOfSmallestOrders+"\n")
	fmt.Fprintf(writer, statsFor8HrPeriods+"\n")
	fmt.Fprintf(writer, exitProgram+"\n")
}

func PrintTable(writer io.Writer, orders []models.Order) {
	fmt.Fprintf(writer, "\nId	Date and time of order	Order type	"+
		"Pay amount		Currency 	Exchange rate	\n")
	for _, order := range orders {
		fmt.Fprintf(writer, "%d %s %s %f %s %f\n", order.Id, order.TimeStamp, order.Type, order.Amount,
			order.Currency, order.ExchangeRate)
	}
}

func PrintBiggestOrders(writer io.Writer, orders []models.BiggestOrders) {
	fmt.Fprintf(writer, "\n\tDate\t\tAmount\n")
	for _, order := range orders {
		fmt.Fprintf(writer, "%s   %f\n", order.Date.Format("2006-01-02"), order.TotalUah)
	}
}

func PrintStats(writer io.Writer, stats []models.PeriodStats) {
	fmt.Fprintf(writer, "\nTime period\t\t Total\t   Big\t  Small\n")
	for _, stat := range stats {
		fmt.Fprintf(writer, "%s\t%5d\t%5d\t%5d\n", stat.TimePeriod, stat.TotalSales, stat.BigSales, stat.SmallSales)
	}
}

func TakeInput(writer io.Writer, reader io.Reader, instruction string) (string, error) {
	var value string

	fmt.Fprintf(writer, instruction)
	_, err := fmt.Fscan(reader, &value)
	if err != nil {
		return "", fmt.Errorf("failed to scan value: %w", err)
	}

	return value, nil
}
