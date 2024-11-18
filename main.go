package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/bruce-mig/stock-exchange/client"
	mm "github.com/bruce-mig/stock-exchange/marketmaker"
	"github.com/bruce-mig/stock-exchange/server"
)

func main() {
	//if go code crashes, we get file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	go server.StartServer()

	//wait for server to boot up b4 client send reqq
	time.Sleep(1 * time.Second)

	c := client.NewClient()

	cfg := mm.Config{
		UserID:         8,
		OrderSize:      200, //original 10
		MinSpread:      20,
		MakeInterval:   1 * time.Second,
		SeedOffset:     40,
		ExchangeClient: c,
		PriceOffset:    10,
	}

	maker := mm.NewMarketMaker(cfg)

	maker.Start()

	time.Sleep(2 * time.Second)
	go marketOrderPlacer(c)

	select {}
}

func marketOrderPlacer(c *client.Client) {
	ticker := time.NewTicker(500 * time.Millisecond)

	for {
		randInt := rand.Intn(10)
		bid := true
		if randInt < 7 {
			bid = false
		}

		order := &client.PlaceOrderParams{
			UserID: 7,
			Bid:    bid,
			Size:   10,
		}

		_, err := c.PlaceMarketOrder(order)
		if err != nil {
			panic(err)
		}

		<-ticker.C
	}
}
