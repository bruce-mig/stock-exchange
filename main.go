package main

import (
	"time"

	"github.com/bruce-mig/stock-exchange/client"
	"github.com/bruce-mig/stock-exchange/server"
)

func main() {
	go server.StartServer()

	time.Sleep(1 * time.Second)

	c := client.NewClient()

	for {
		limitOrderParams1 := &client.PlaceOrderParams{
			UserID: 8,
			Bid:    false,
			Price:  10_000,
			Size:   500_000,
		}

		_, err := c.PlaceLimitOrder(limitOrderParams1)
		if err != nil {
			panic(err)
		}

		limitOrderParams2 := &client.PlaceOrderParams{
			UserID: 6,
			Bid:    false,
			Price:  9_000,
			Size:   500_000,
		}

		_, err = c.PlaceLimitOrder(limitOrderParams2)
		if err != nil {
			panic(err)
		}

		// fmt.Println("placed limit order from the client =>", res.OrderID)

		marketOrderParams := &client.PlaceOrderParams{
			UserID: 7,
			Bid:    true,
			Size:   1_000_000,
		}

		_, err = c.PlaceMarketOrder(marketOrderParams)
		if err != nil {
			panic(err)
		}

		// fmt.Println("placed market order from the client =>", res.OrderID)
		time.Sleep(1 * time.Second)
	}

	//for blocking
	select {}
}
