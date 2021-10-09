package main

import (
	"C"
	"context"
	"github.com/adshao/go-binance/v2"
	"github.com/joho/godotenv"
	"log"
	"math"
	"os"
	"strconv"
)

type BookSide int
type AmountType int
type OrderMode int

const (
	Bid BookSide = iota
	Ask
)

const (
	FixedBase AmountType = iota
	FixedQuote
)

const (
	Live OrderMode = iota
	Test
)

// OrderCommand - used to place an order on BinanceUs
type OrderCommand struct {
	Symbol   string
	Side     binance.SideType
	Quantity string
	Price    string
	Client   *binance.Client
	Mode     OrderMode
}

// OrderArgs - Conversion from Python input to struct
type OrderArgs struct {
	Symbol            string
	Side              int
	AmountType        int
	Amount            string
	QuantityPrecision int
	Mode              int
}

func getClient() *binance.Client {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

    apiKey := os.Getenv("API_KEY")
    secret := os.Getenv("SECRET")

    client := binance.NewClient(apiKey, secret)
	client.BaseURL = "https://api.binance.us"

	return client
}

func getTicker(client *binance.Client, symbol string) *binance.BookTicker {
	tickers, err := client.NewListBookTickersService().Symbol(symbol).Do(context.Background())

	if err != nil {
		log.Fatal(err)
	}

	return tickers[0]
}

func getBookPrice(ticker *binance.BookTicker, side BookSide) string {
	switch side {
	case Bid:
		return ticker.BidPrice
	case Ask:
		return ticker.AskPrice
	}
	return ""
}

func roundDown(input float64, places int) float64 {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * input
	round = math.Floor(digit)
	return round / pow
}

func amountFromQuote(price string, amount string, precision int) (string, error) {
	priceFloat, err := strconv.ParseFloat(price, 64)
	amountFloat, err := strconv.ParseFloat(amount, 64)

	if err != nil {
		return "", err
	}

	rounded := roundDown(amountFloat/priceFloat, precision)
	// Strip any trailing zeros
	return strconv.FormatFloat(rounded, 'f', -1, 64), nil
}

func quantityFromBase(amount string, precision int) (string, error) {
	amountFloat, err := strconv.ParseFloat(amount, 64)

	if err != nil {
		return "", err
	}

	rounded := roundDown(amountFloat, precision)
	return strconv.FormatFloat(rounded, 'f', -1, 64), nil
}

func (orderArgs *OrderArgs) CreateOrder() OrderCommand {
	var orderSide binance.SideType
	var bookSide BookSide
	var amtType AmountType
	var quantity string
	var mode OrderMode
	var err error

	switch orderArgs.Side {
	case 0:
		orderSide = binance.SideTypeBuy
		bookSide = Bid
	default:
		orderSide = binance.SideTypeSell
		bookSide = Ask
	}

	switch orderArgs.AmountType {
	case 0:
		amtType = FixedBase
	default:
		amtType = FixedQuote
	}

	switch orderArgs.Mode {
	case 0:
		mode = Live
	default:
		mode = Test
	}

	client := getClient()
	ticker := getTicker(client, orderArgs.Symbol)
	price := getBookPrice(ticker, bookSide)

	if amtType == FixedBase {
		//quantity = orderArgs.Amount
		quantity, err = quantityFromBase(orderArgs.Amount, orderArgs.QuantityPrecision)

		if err != nil {
			log.Fatal(err)
		}
	} else {
		quantity, err = amountFromQuote(price, orderArgs.Amount, orderArgs.QuantityPrecision)

		if err != nil {
			log.Fatal(err)
		}
	}

	return OrderCommand{
		Symbol:   orderArgs.Symbol,
		Side:     orderSide,
		Quantity: quantity,
		Price:    price,
		Client:   client,
		Mode:     mode,
	}
}

// testOrder - test that an order is valid without placing.
func (order *OrderCommand) testOrder() (int, error) {
	res := order.Client.NewCreateOrderService().Symbol(order.Symbol).
		Side(order.Side).Type(binance.OrderTypeLimit).
		TimeInForce(binance.TimeInForceTypeGTC).Quantity(order.Quantity).
		Price(order.Price).Test(context.Background())

	log.Printf("Order Response:\t%v", res)
	return 0, res
	//return "", res
}

// executeOrder - make sure that this returns the order id to Python
func (order *OrderCommand) executeOrder() (int, error) {
	res, err := order.Client.NewCreateOrderService().Symbol(order.Symbol).
		Side(order.Side).Type(binance.OrderTypeLimit).
		TimeInForce(binance.TimeInForceTypeGTC).Quantity(order.Quantity).
		Price(order.Price).Do(context.Background())

	if err != nil {
		log.Println(err)
		return 0, err
	}
	log.Printf("Order Response:\t%v\n", res)
	idInt := int(res.OrderID)
	return idInt, nil
}

func ParseArgs(symbol *C.char, side *C.char, amountType *C.char, amount *C.char, precision *C.char, mode *C.char) OrderArgs {
	symbolStr := C.GoString(symbol)
	sideStr := C.GoString(side)
	amountTypeStr := C.GoString(amountType)
	amountStr := C.GoString(amount)
	precisionStr := C.GoString(precision)
	modeStr := C.GoString(mode)

	sideInt, err := strconv.Atoi(sideStr)
	amountTypeInt, err := strconv.Atoi(amountTypeStr)
	precisionInt, err := strconv.Atoi(precisionStr)
	modeInt, err := strconv.Atoi(modeStr)

	if err != nil {
		log.Fatal(err)
	}

	return OrderArgs{
		Symbol:            symbolStr,
		Side:              sideInt,
		AmountType:        amountTypeInt,
		Amount:            amountStr,
		QuantityPrecision: precisionInt,
		Mode:              modeInt,
	}
}

//export PlaceOrder
func PlaceOrder(symbol *C.char, side C.int, amountType C.int, amount *C.char, precision C.int, mode C.int) C.int {
	var orderId int
	var err error

	sym := C.GoString(symbol)
	amt := C.GoString(amount)
	args := OrderArgs{
		Symbol:            sym,
		Side:              int(side),
		AmountType:        int(amountType),
		Amount:            amt,
		QuantityPrecision: int(precision),
		Mode:              int(mode),
	}

	cmd := args.CreateOrder()

	if cmd.Mode == Live {
		orderId, err = cmd.executeOrder()

		if err != nil {
			log.Println(err)
			orderId = 0
		}
	} else {
		orderId, err = cmd.testOrder()

		if err != nil {
			log.Println(err)
			orderId = 0
		}
	}
	return C.int(orderId)
}

//export LowBuy
func LowBuy(symbol *C.char, side C.int, amountType C.int, amount *C.char, precision C.int, mode C.int, pricePrecision C.int) C.int {
	/*
		This places an order 20% below the bid price. The purpose is to provide an interface to create a real
		order that is unlikely to fill, while working out application logic that is dependent on testing with
		real data.
		If it does fill, well you get a good deal then.
	*/
	var err error
	var orderId int
	sym := C.GoString(symbol)
	amt := C.GoString(amount)
	args := OrderArgs{
		Symbol:            sym,
		Side:              int(side),
		AmountType:        int(amountType),
		Amount:            amt,
		QuantityPrecision: int(precision),
		Mode:              int(mode),
	}

	if args.Side != 0 {
		args.Side = 0
	}

	pricePrec := int(pricePrecision)

	cmd := args.CreateOrder()

	priceFloat, err := strconv.ParseFloat(cmd.Price, 64)

	if err != nil {
		panic(err)
	}

	newPriceFloat := priceFloat * 0.80
	log.Printf("New Calculated Price:\t%v\n", newPriceFloat)

	newPriceRounded := roundDown(newPriceFloat, pricePrec)

	strPrice := strconv.FormatFloat(newPriceRounded, 'f', -1, 64)
	log.Printf("Price as string:\t%v\n", strPrice)

	updatedCmd := OrderCommand{
		Symbol:   cmd.Symbol,
		Quantity: cmd.Quantity,
		Price:    strPrice,
		Side:     binance.SideTypeBuy,
		Mode:     cmd.Mode,
		Client:   cmd.Client,
	}

	if updatedCmd.Mode == Live {
		orderId, err = updatedCmd.executeOrder()

		if err != nil {
			log.Println(err)
			orderId = 0
		}
	} else {
		orderId, err = updatedCmd.testOrder()

		if err != nil {
			log.Println(err)
			orderId = 0
		}
	}

	log.Printf("Order Id in Final Function:\t%v\n", orderId)
	return C.int(orderId)
}

//export HighSell
func HighSell(symbol *C.char, side C.int, amountType C.int, amount *C.char, precision C.int, mode C.int, pricePrecision C.int) C.int {
	var err error
	var orderId int

	sym := C.GoString(symbol)
	amt := C.GoString(amount)
	args := OrderArgs{
		Symbol:            sym,
		Side:              int(side),
		AmountType:        int(amountType),
		Amount:            amt,
		QuantityPrecision: int(precision),
		Mode:              int(mode),
	}

	if args.Side != 1 {
		args.Side = 1
	}

	pricePrec := int(pricePrecision)

	cmd := args.CreateOrder()

	priceFloat, err := strconv.ParseFloat(cmd.Price, 64)

	if err != nil {
		panic(err)
	}

	highPrice := priceFloat * 1.20

	rounded := roundDown(highPrice, pricePrec)
	newPrice := strconv.FormatFloat(rounded, 'f', -1, 64)

	updatedCmd := OrderCommand{
		Symbol:   cmd.Symbol,
		Quantity: cmd.Quantity,
		Price:    newPrice,
		Side:     binance.SideTypeSell,
		Mode:     cmd.Mode,
		Client:   cmd.Client,
	}

	if updatedCmd.Mode == Live {
		orderId, err = updatedCmd.executeOrder()

		if err != nil {
			log.Println(err)
			orderId = 0
		}
	} else {
		orderId, err = updatedCmd.testOrder()

		if err != nil {
			log.Println(err)
			orderId = 0
		}
	}
	return C.int(orderId)
}

func CustomPriceOrder(orderArgs *OrderArgs, price string, pricePrecision int) OrderCommand {
	var orderSide binance.SideType
	var amtType AmountType
	var quantity string
	var mode OrderMode
	var err error

	switch orderArgs.Side {
	case 0:
		orderSide = binance.SideTypeBuy
	default:
		orderSide = binance.SideTypeSell
	}

	switch orderArgs.AmountType {
	case 0:
		amtType = FixedBase
	default:
		amtType = FixedQuote
	}

	switch orderArgs.Mode {
	case 0:
		mode = Live
	default:
		mode = Test
	}

	client := getClient()
	priceFloat, err := strconv.ParseFloat(price, 64)

	if err != nil {
		panic(err)
	}

	priceRounded := roundDown(priceFloat, pricePrecision)
	newPrice := strconv.FormatFloat(priceRounded, 'f', -1, 64)
	log.Printf("Execution Price:\t%v\n", newPrice)

	if amtType == FixedBase {
		quantity, err = quantityFromBase(orderArgs.Amount, orderArgs.QuantityPrecision)
		log.Printf("Quantity:\t%v\n", quantity)

		if err != nil {
			log.Fatal(err)
		}
	} else {
		quantity, err = amountFromQuote(newPrice, orderArgs.Amount, orderArgs.QuantityPrecision)
		log.Printf("Quantity:\t%v\n", quantity)

		if err != nil {
			log.Fatal(err)
		}
	}

	return OrderCommand{
		Symbol:   orderArgs.Symbol,
		Side:     orderSide,
		Quantity: quantity,
		Price:    newPrice,
		Client:   client,
		Mode:     mode,
	}
}

//export TradeCustomPrice
func TradeCustomPrice(symbol *C.char, side C.int, amountType C.int,
	amount *C.char, precision C.int, mode C.int, price *C.char, pricePrecision C.int) C.int {
	var err error
	var orderId int
	var goSymbol string = C.GoString(symbol)
	goSide := int(side)
	goAmountType := int(amountType)
	var goAmount string = C.GoString(amount)
	goPrecision := int(precision)
	goMode := int(mode)
	var goPrice string = C.GoString(price)
	goPricePrecision := int(pricePrecision)

	args := OrderArgs{
		Symbol:            goSymbol,
		Side:              goSide,
		AmountType:        goAmountType,
		Amount:            goAmount,
		QuantityPrecision: goPrecision,
		Mode:              goMode,
	}

	cmd := CustomPriceOrder(&args, goPrice, goPricePrecision)

	if cmd.Mode == Live {
		orderId, err = cmd.executeOrder()

		if err != nil {
			log.Println(err)
			orderId = 0
		}
	} else {
		orderId, err = cmd.testOrder()

		if err != nil {
			log.Println(err)
			orderId = 0
		}
	}

	return C.int(orderId)
}

func main() {
}
