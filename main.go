package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/bruce-mig/stock-exchange/client"
	"github.com/bruce-mig/stock-exchange/server"
)

var (
	tick   = 5 * time.Second
	myAsks = make(map[float64]int64) //[price]orderID
	myBids = make(map[float64]int64)
)

const (
	maxOrders = 3
)

func marketOrderPlacer(c *client.Client) {
	ticker := time.NewTicker(tick)
	for {

		marketSellOrder := &client.PlaceOrderParams{
			UserID: 6,
			Bid:    false,
			Size:   1000,
		}

		orderResp, err := c.PlaceMarketOrder(marketSellOrder)
		if err != nil {
			log.Println(orderResp.OrderID)
		}

		marketBuyOrder := &client.PlaceOrderParams{
			UserID: 6,
			Bid:    true,
			Size:   1000,
		}

		orderResp, err = c.PlaceMarketOrder(marketBuyOrder)
		if err != nil {
			log.Println(orderResp.OrderID)
		}

		<-ticker.C
	}
}

func makeMarketSimple(c *client.Client) {
	ticker := time.NewTicker(tick)
	for {

		orders, err := c.GetOrders(7)
		if err != nil {
			log.Println(err)
		}

		fmt.Printf("------------------------------------------\n")
		fmt.Printf("%+v\n", orders)
		fmt.Printf("------------------------------------------\n")

		bestAsk, err := c.GetBestAsk()
		if err != nil {
			log.Println(err)
		}

		bestBid, err := c.GetBestBid()
		if err != nil {
			log.Println(err)
		}

		spread := math.Abs(bestBid - bestAsk)
		fmt.Println("exchange spread:", spread)

		if len(myBids) < maxOrders {
			// place the bid
			bidLimit := &client.PlaceOrderParams{
				UserID: 7,
				Bid:    true,
				Price:  bestBid + 100, //add straddle
				Size:   1000,
			}

			bidOrderResp, err := c.PlaceLimitOrder(bidLimit)
			if err != nil {
				log.Println(bidOrderResp.OrderID)
			}

			myBids[bidLimit.Price] = bidOrderResp.OrderID

		}

		if len(myAsks) < maxOrders {
			// place the ask
			askLimit := &client.PlaceOrderParams{
				UserID: 7,
				Bid:    false,
				Price:  bestAsk - 100, //less straddle to tighten spread
				Size:   1000,
			}

			askOrderResp, err := c.PlaceLimitOrder(askLimit)
			if err != nil {
				log.Println(askOrderResp.OrderID)
			}

			myAsks[askLimit.Price] = askOrderResp.OrderID
		}

		fmt.Println("best ask price", bestAsk)
		fmt.Println("best bid price", bestBid)

		<-ticker.C
	}
}

func seedMarket(c *client.Client) error {
	ask := &client.PlaceOrderParams{
		UserID: 8,
		Bid:    false,
		Price:  10_000,
		Size:   1_000_000,
	}

	bid := &client.PlaceOrderParams{
		UserID: 8,
		Bid:    true,
		Price:  9_000,
		Size:   1_000_000,
	}

	_, err := c.PlaceLimitOrder(ask)
	if err != nil {
		return err
	}

	_, err = c.PlaceLimitOrder(bid)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	go server.StartServer()

	//wait for server to boot up b4 client send reqq
	time.Sleep(1 * time.Second)

	c := client.NewClient()

	if err := seedMarket(c); err != nil {
		panic(err)
	}

	go makeMarketSimple(c)
	time.Sleep(1 * time.Second) // prevent blocking marketOrderPlacer
	marketOrderPlacer(c)

	// for {
	// 	limitOrderParams1 := &client.PlaceOrderParams{
	// 		UserID: 8,
	// 		Bid:    false,
	// 		Price:  10_000,
	// 		Size:   5_000_000,
	// 	}

	// 	_, err := c.PlaceLimitOrder(limitOrderParams1)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	limitOrderParams2 := &client.PlaceOrderParams{
	// 		UserID: 6,
	// 		Bid:    false,
	// 		Price:  9_000,
	// 		Size:   500_000,
	// 	}

	// 	_, err = c.PlaceLimitOrder(limitOrderParams2)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	// fmt.Println("placed limit order from the client =>", res.OrderID)

	// 	buyLimitOrder := &client.PlaceOrderParams{
	// 		UserID: 6,
	// 		Bid:    true,
	// 		Price:  11_000,
	// 		Size:   500_000,
	// 	}

	// 	_, err = c.PlaceLimitOrder(buyLimitOrder)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	marketOrderParams := &client.PlaceOrderParams{
	// 		UserID: 7,
	// 		Bid:    true,
	// 		Size:   1_000_000,
	// 	}

	// 	_, err = c.PlaceMarketOrder(marketOrderParams)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	bestBidPrice, err := c.GetBestBid()
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	fmt.Println("best bid price", bestBidPrice)

	// 	bestAskPrice, err := c.GetBestAsk()
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	fmt.Println("best ask price", bestAskPrice)

	// 	// fmt.Println("placed market order from the client =>", res.OrderID)
	// 	time.Sleep(1 * time.Second)
	// }

	//for blocking
	select {}
}
