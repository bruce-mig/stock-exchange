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
	tick = 5 * time.Second
	// myAsks = make(map[float64]int64) //[price]orderID
	// myBids = make(map[float64]int64)
)

const (
	maxOrders = 3
	userID    = 7
)

func marketOrderPlacer(c *client.Client) {
	ticker := time.NewTicker(tick)
	for {

		trades, err := c.GetTrades("INN")
		if err != nil {
			panic(err)
		}

		if len(trades) > 0 {
			fmt.Printf("EXCHANGE PRICE:====>%.2f\n", trades[len(trades)-1].Price)
		}

		otherMarketSellOrder := &client.PlaceOrderParams{
			UserID: 8,
			Bid:    false,
			Size:   1000,
		}

		orderResp, err := c.PlaceMarketOrder(otherMarketSellOrder)
		if err != nil {
			log.Println(orderResp.OrderID)
		}

		marketSellOrder := &client.PlaceOrderParams{
			UserID: 6,
			Bid:    false,
			Size:   100,
		}

		orderResp, err = c.PlaceMarketOrder(marketSellOrder)
		if err != nil {
			log.Println(orderResp.OrderID)
		}

		marketBuyOrder := &client.PlaceOrderParams{
			UserID: 6,
			Bid:    true,
			Size:   100,
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

		orders, err := c.GetOrders(userID)
		if err != nil {
			log.Println(err)
		}

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

		if len(orders.Bids) < maxOrders {
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

			// myBids[bidLimit.Price] = bidOrderResp.OrderID

		}

		if len(orders.Asks) < maxOrders {
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

			// myAsks[askLimit.Price] = askOrderResp.OrderID
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
		Size:   100_000,
	}

	bid := &client.PlaceOrderParams{
		UserID: 8,
		Bid:    true,
		Price:  9_000,
		Size:   100_000,
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
	//if go code crashes, we get file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

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

	select {}
}
