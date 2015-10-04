package main

import (
	"github.com/bitly/go-simplejson"
	"os"
	"strconv"
	"strings"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	if len(os.Args) < 2 || len(os.Args) > 4 {
		fmt.Println("Number of arguments is invalid!")
		howTo()
		return
	} else if len(os.Args) == 2 { 
		_, err := strconv.ParseInt(os.Args[1], 10, 64)
		if err != nil {
			fmt.Println("Illegal argument/s!")
			howTo()
			return
		}
		data, err := json.Marshal(map[string]interface{}{
			"method": "StockAccounts.Check",
			"id":     1,
			"params": []map[string]interface{}{map[string]interface{}{"TradeId": os.Args[1]}},
		})
		if err != nil {
			log.Fatalf("Marshal : %v", err)
		}
		resp, err := http.Post("http://127.0.0.1:5062/rpc", "application/json", strings.NewReader(string(data)))
		if err != nil {
			log.Fatalf("Post: %v", err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("ReadAll: %v", err)
		}
		newjson, err := simplejson.NewJson(body)
		checkError(err)
		fmt.Print("Stocks: ")
		stocks := newjson.Get("result").Get("Stocks")
		fmt.Println(stocks)
		fmt.Print("Uninvested Amount: ")
		uninvestedAmount, _ := newjson.Get("result").Get("UninvestedAmount").Float64()
		fmt.Print("$")
		fmt.Println(uninvestedAmount)
		fmt.Print("Total Market Value: ")
		totalMarketValue, _ := newjson.Get("result").Get("TotalMarketValue").Float64()
		fmt.Print("$")
		fmt.Println(totalMarketValue)
	} else if len(os.Args) == 3 { 
		budget, err := strconv.ParseFloat(os.Args[2], 64)
		if err != nil {
			fmt.Println("Wrong argument: budget")
			howTo()
			return
		}
		data, err := json.Marshal(map[string]interface{}{
			"method": "StockAccounts.Buy",
			"id":     2,
			"params": []map[string]interface{}{map[string]interface{}{"StockSymbolAndPercentage": os.Args[1], "Budget": float32(budget)}},
		})
		if err != nil {
			log.Fatalf("Marshal : %v", err)
		}
		resp, err := http.Post("http://127.0.0.1:5062/rpc", "application/json", strings.NewReader(string(data)))
		if err != nil {
			log.Fatalf("Post: %v", err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("ReadAll: %v", err)
		}
		newjson, err := simplejson.NewJson(body)
		checkError(err)
		fmt.Print("Trade ID: ")
		tradeId, _ := newjson.Get("result").Get("TradeId").Int()
		fmt.Println(tradeId)
		fmt.Print("Stocks: ")
		stocks := newjson.Get("result").Get("Stocks")
		fmt.Println(*stocks)
		fmt.Print("Uninvested Amount: ")
		uninvestedAmount, _ := newjson.Get("result").Get("UninvestedAmount").Float64()
		fmt.Print("$")
		fmt.Println(uninvestedAmount)
	} else {
		fmt.Println("Error!")
		howTo()
		return
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s\n", err.Error())
		log.Fatal("error: ", err)
		os.Exit(2)
	}
}

func howTo() {
	fmt.Println("Usage: ", os.Args[0], "tradeId")
	fmt.Println("or")
	fmt.Println(os.Args[0], "“GOOG:40%,YHOO:60%” 12500")
}