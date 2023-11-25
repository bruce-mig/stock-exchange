package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bruce-mig/stock-exchange/server"
)

const Endpoint = "http://localhost:3000"

type (
	Client struct {
		*http.Client
	}

	PlaceOrderParams struct {
		UserID int64
		Bid    bool
		// Price only needed for placing LIMIT orders
		Price float64
		Size  float64
	}
)

func NewClient() *Client {
	return &Client{
		Client: http.DefaultClient,
	}
}

func (c *Client) PlaceMarketOrder(p *PlaceOrderParams) (*server.PlaceOrderResponse, error) {
	params := &server.PlaceOrderRequest{
		UserID: p.UserID,
		Type:   server.MarketOrder,
		Bid:    p.Bid,
		Size:   p.Size,
		Market: server.MarketINN,
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	e := Endpoint + "/order"
	req, err := http.NewRequest(http.MethodPost, e, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	placeOrderResponse := &server.PlaceOrderResponse{}
	if err := json.NewDecoder(res.Body).Decode(placeOrderResponse); err != nil {
		return nil, err
	}
	return placeOrderResponse, nil
}

func (c *Client) GetBestBid() (float64, error) {
	e := fmt.Sprintf("%s/book/%s/bestBid", Endpoint, server.MarketINN)

	req, err := http.NewRequest(http.MethodGet, e, nil)
	if err != nil {
		return 0, err
	}

	res, err := c.Do(req)
	if err != nil {
		return 0, err
	}

	priceResp := &server.PriceResponse{}
	if err := json.NewDecoder(res.Body).Decode(priceResp); err != nil {
		return 0, err
	}
	return priceResp.Price, err
}

func (c *Client) GetBestAsk() (float64, error) {
	e := fmt.Sprintf("%s/book/%s/bestAsk", Endpoint, server.MarketINN)

	req, err := http.NewRequest(http.MethodGet, e, nil)
	if err != nil {
		return 0, err
	}

	res, err := c.Do(req)
	if err != nil {
		return 0, err
	}

	priceResp := &server.PriceResponse{}
	if err := json.NewDecoder(res.Body).Decode(priceResp); err != nil {
		return 0, err
	}
	return priceResp.Price, err
}

func (c *Client) CancelOrder(orderID int64) error {
	e := fmt.Sprintf("%s/order/%d", Endpoint, orderID)
	req, err := http.NewRequest(http.MethodDelete, e, nil)
	if err != nil {
		return err
	}

	_, err = c.Do(req)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) PlaceLimitOrder(p *PlaceOrderParams) (*server.PlaceOrderResponse, error) {
	params := &server.PlaceOrderRequest{
		UserID: p.UserID,
		Type:   server.LimitOrder,
		Bid:    p.Bid,
		Size:   p.Size,
		Price:  p.Price,
		Market: server.MarketINN,
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	e := Endpoint + "/order"
	req, err := http.NewRequest(http.MethodPost, e, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	placeOrderResponse := &server.PlaceOrderResponse{}
	if err := json.NewDecoder(res.Body).Decode(placeOrderResponse); err != nil {
		return nil, err
	}
	return placeOrderResponse, nil
}
