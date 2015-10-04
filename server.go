package main
import (
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
	"math"
	"github.com/bakins/net-http-recover"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"github.com/justinas/alice"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type StockAccounts struct {
	stockPortfolio map[int](*Portfolio)
}

type Portfolio struct {
	stocks          map[string](*Share)
	UninvestedAmount 	float32
}

type Share struct {
	boughtPrice float32
	shareId    int
}

type BuyRequest struct {
	StockSymbolAndPercentage string
	Budget                   float32
}

type BuyResponse struct {
	TradeId         int
	Stocks           []string
	UninvestedAmount float32
}

type CheckRequest struct {
	TradeId string
}

type CheckResponse struct {
	Stocks           []string
	UninvestedAmount float32
	TotalMarketValue float32
}

var st StockAccounts

var tradeId int

func main() {

	var st = (new(StockAccounts))

	tradeId = 100

	router := mux.NewRouter()
	server := rpc.NewServer()
	server.RegisterCodec(json.NewCodec(), "application/json")
	server.RegisterService(st, "")

	chain := alice.New(
		func(h http.Handler) http.Handler {
			return handlers.CombinedLoggingHandler(os.Stdout, h)
		},
		handlers.CompressHandler,
		func(h http.Handler) http.Handler {
			return recovery.Handler(os.Stderr, h, true)
		})

	router.Handle("/rpc", chain.Then(server))
	log.Fatal(http.ListenAndServe(":5062", server))
}

func (st *StockAccounts) Buy(httpRequest *http.Request, request *BuyRequest, response *BuyResponse) error {

	tradeId++
	response.TradeId = tradeId

	if st.stockPortfolio == nil {

		st.stockPortfolio = make(map[int](*Portfolio))

		st.stockPortfolio[tradeId] = new(Portfolio)
		st.stockPortfolio[tradeId].stocks = make(map[string]*Share)

	}

	symbolAndPercentages := strings.Split(request.StockSymbolAndPercentage, ",")
	newbudget := float32(request.Budget)
	var spent float32

	for _, stk := range symbolAndPercentages {

		split := strings.Split(stk, ":")
		stockQuote := split[0]
		percentage := split[1]
		strPercentage := strings.TrimSuffix(percentage, "%")
		floatPercentage64, _ := strconv.ParseFloat(strPercentage, 32)
		floatPercentage := float32(floatPercentage64 / 100.00)
		currentPrice := checkQuote(stockQuote)

		shares := int(math.Floor(float64(newbudget * floatPercentage / currentPrice)))
		sharesFloat := float32(shares)
		spent += sharesFloat * currentPrice

		if _, ok := st.stockPortfolio[tradeId]; !ok {

			newPortfolio := new(Portfolio)
			newPortfolio.stocks = make(map[string]*Share)
			st.stockPortfolio[tradeId] = newPortfolio
		}
		if _, ok := st.stockPortfolio[tradeId].stocks[stockQuote]; !ok {

			newShare := new(Share)
			newShare.boughtPrice = currentPrice
			newShare.shareId = shares
			st.stockPortfolio[tradeId].stocks[stockQuote] = newShare
		} else {

			total := float32(sharesFloat*currentPrice) + float32(st.stockPortfolio[tradeId].stocks[stockQuote].shareId)*st.stockPortfolio[tradeId].stocks[stockQuote].boughtPrice
			st.stockPortfolio[tradeId].stocks[stockQuote].boughtPrice = total / float32(shares+st.stockPortfolio[tradeId].stocks[stockQuote].shareId)
			st.stockPortfolio[tradeId].stocks[stockQuote].shareId += shares
		}

		stockBought := stockQuote + ":" + strconv.Itoa(shares) + ":$" + strconv.FormatFloat(float64(currentPrice), 'f', 2, 32)

		response.Stocks = append(response.Stocks, stockBought)
	}

	leftOver := newbudget - spent
	response.UninvestedAmount = leftOver
	st.stockPortfolio[tradeId].UninvestedAmount += leftOver

	return nil
}

func (st *StockAccounts) Check(httpRequest *http.Request, checkRq *CheckRequest, checkResp *CheckResponse) error {

	if st.stockPortfolio == nil {
		return errors.New("No account set up yet.")
	}

	tradeNum64, err := strconv.ParseInt(checkRq.TradeId, 10, 64)

	if err != nil {
		return errors.New("Illegal Trade ID. ")
	}
	tradeId := int(tradeNum64)

	if pocket, ok := st.stockPortfolio[tradeId]; ok {

		var currentMarketVal float32
		for stockquote, sh := range pocket.stocks {
			currentPrice := checkQuote(stockquote)

			var str string
			if sh.boughtPrice < currentPrice {
				str = "+$" + strconv.FormatFloat(float64(currentPrice), 'f', 2, 32)
			} else if sh.boughtPrice > currentPrice {
				str = "-$" + strconv.FormatFloat(float64(currentPrice), 'f', 2, 32)
			} else {
				str = "$" + strconv.FormatFloat(float64(currentPrice), 'f', 2, 32)
			}

			entry := stockquote + ":" + strconv.Itoa(sh.shareId) + ":" + str

			checkResp.Stocks = append(checkResp.Stocks, entry)

			currentMarketVal += float32(sh.shareId) * currentPrice
		}

		checkResp.UninvestedAmount = pocket.UninvestedAmount

		checkResp.TotalMarketValue = currentMarketVal
	} else {
		return errors.New("No such trade ID. ")
	}

	return nil
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func checkQuote(stockName string) float32 {
	leftBaseUrl := "https://query.yahooapis.com/v1/public/yql?q=select%20LastTradePriceOnly%20from%20yahoo.finance%0A.quotes%20where%20symbol%20%3D%20%22"
	rightBaseUrl := "%22%0A%09%09&format=json&env=http%3A%2F%2Fdatatables.org%2Falltables.env"
	resp, err := http.Get(leftBaseUrl + stockName + rightBaseUrl)

	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != 200 {
		log.Fatal("Query failure, possibly no network connection or illegal stock quote ")
	}

	newjson, err := simplejson.NewJson(body)
	if err != nil {
		fmt.Println(err)
	}

	price, _ := newjson.Get("query").Get("results").Get("quote").Get("LastTradePriceOnly").String()
	floatPrice, err := strconv.ParseFloat(price, 32)
	return float32(floatPrice)
}
