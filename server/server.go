package server

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/bruce-mig/stock-exchange/orderbook"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"

	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
)

const (
	MarketOrder OrderType = "MARKET"
	LimitOrder  OrderType = "LIMIT"
	MarketINN   Market    = "INN.ZSX"
)

var (
	exchangePrivateKey = os.Getenv("EXCHANGE_PK")
	csdEndpoint        = os.Getenv("CSD_ENDPOINT")
)

type (
	OrderType string
	Market    string

	Exchange struct {
		Client *ethclient.Client
		mu     sync.RWMutex
		Users  map[string]*User
		// Orders maps a user to his orders.
		Orders     map[string][]*orderbook.Order
		PrivateKey *ecdsa.PrivateKey
		orderbooks map[Market]*orderbook.Orderbook
	}

	PlaceOrderRequest struct {
		UserID string
		Type   OrderType //limit or market
		Bid    bool
		Size   float64
		Price  float64
		Market Market
	}

	Order struct {
		UserID    string
		ID        int64
		Price     float64
		Size      float64
		Bid       bool
		Timestamp int64
	}

	OrderbookData struct {
		TotalBidVolume float64
		TotalAskVolume float64
		Asks           []*Order
		Bids           []*Order
	}

	MatchedOrder struct {
		UserID string
		Price  float64
		Size   float64
		ID     int64
	}

	User struct {
		ID         string
		PrivateKey *ecdsa.PrivateKey
	}

	PlaceOrderResponse struct {
		OrderID int64
	}

	GetOrdersResponse struct {
		Asks []Order
		Bids []Order
	}

	PriceResponse struct {
		Price float64
	}

	APIError struct {
		Error string
	}
)

func StartServer() {
	e := echo.New()
	e.HTTPErrorHandler = httpErrorHandler

	client, err := ethclient.Dial(csdEndpoint)
	if err != nil {
		log.Fatal(err)
	}
	ex, err := NewExchange(exchangePrivateKey, client)
	if err != nil {
		log.Fatal(err)
	}

	ex.registerUser(os.Getenv("USER_1_PK"), "CSD000000000001-0001")
	ex.registerUser(os.Getenv("USER_2_PK"), "CSD000000000002-0001")
	ex.registerUser(os.Getenv("USER_3_PK"), "CSD000000000003-0001")

	e.GET("/trades/:market", ex.handleGetTrades)
	e.GET("/order/user/:userID", ex.handleGetOrders)
	e.GET("/book/:market", ex.handleGetBook)
	e.GET("/book/:market/bestBid", ex.handleGetBestBid)
	e.GET("/book/:market/bestAsk", ex.handleGetBestAsk)

	e.POST("/order", ex.handlePlaceOrder)
	e.DELETE("/order/:id", ex.cancelOrder)

	e.Start(":3000")
}

func NewUser(privKey string, id string) *User {
	pk, err := crypto.HexToECDSA(privKey)
	if err != nil {
		panic(err)
	}
	return &User{
		ID:         id,
		PrivateKey: pk,
	}
}

func httpErrorHandler(err error, c echo.Context) {
	fmt.Println(err)
}

func NewExchange(privateKey string, client *ethclient.Client) (*Exchange, error) {
	orderbooks := make(map[Market]*orderbook.Orderbook)
	orderbooks[MarketINN] = orderbook.NewOrderbook()

	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, err
	}

	return &Exchange{
		Client:     client,
		Users:      make(map[string]*User),
		Orders:     make(map[string][]*orderbook.Order),
		PrivateKey: pk,
		orderbooks: orderbooks,
	}, nil
}

func (ex *Exchange) handleGetTrades(c echo.Context) error {
	market := Market(c.Param("market"))
	ob, ok := ex.orderbooks[market]
	if !ok {
		return c.JSON(http.StatusBadRequest, APIError{Error: "orderbook not found"})
	}
	return c.JSON(http.StatusOK, ob.Trades)
}

func (ex *Exchange) handleGetOrders(c echo.Context) error {
	userID := c.Param("userID")

	ex.mu.RLock()

	orderbookOrders := ex.Orders[userID]
	orderResp := &GetOrdersResponse{
		Asks: []Order{},
		Bids: []Order{},
	}

	for i := 0; i < len(orderbookOrders); i++ {
		if orderbookOrders[i].Limit == nil {
			continue
		}

		order := Order{
			UserID:    orderbookOrders[i].UserID,
			ID:        orderbookOrders[i].ID,
			Price:     orderbookOrders[i].Limit.Price,
			Size:      orderbookOrders[i].Size,
			Bid:       orderbookOrders[i].Bid,
			Timestamp: orderbookOrders[i].Timestamp,
		}

		if order.Bid {
			orderResp.Bids = append(orderResp.Bids, order)
		} else {
			orderResp.Asks = append(orderResp.Asks, order)
		}
	}
	ex.mu.RUnlock()

	return c.JSON(http.StatusOK, orderResp)
}

func (ex *Exchange) handleGetBook(c echo.Context) error {
	market := Market(c.Param("market"))
	ob, ok := ex.orderbooks[market]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "market not found"})
	}

	orderbookData := OrderbookData{
		TotalBidVolume: ob.BidTotalVolume(),
		TotalAskVolume: ob.AskTotalVolume(),
		Asks:           []*Order{},
		Bids:           []*Order{},
	}
	for _, limit := range ob.Asks() {
		for _, order := range limit.Orders {
			o := Order{
				UserID:    order.UserID,
				ID:        order.ID,
				Price:     limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			}
			orderbookData.Asks = append(orderbookData.Asks, &o)
		}
	}

	for _, limit := range ob.Bids() {
		for _, order := range limit.Orders {
			o := Order{
				UserID:    order.UserID,
				ID:        order.ID,
				Price:     limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			}
			orderbookData.Bids = append(orderbookData.Bids, &o)
		}
	}
	return c.JSON(http.StatusOK, orderbookData)
}

func (ex *Exchange) handleGetBestBid(c echo.Context) error {

	market := Market(c.Param("market"))
	ob := ex.orderbooks[market]
	order := Order{}

	if len(ob.Bids()) == 0 {
		return c.JSON(http.StatusOK, order)
	}
	bestLimit := ob.Bids()[0]
	bestOrder := bestLimit.Orders[0]

	order.Price = bestLimit.Price
	order.UserID = bestOrder.UserID

	return c.JSON(http.StatusOK, order)

}

func (ex *Exchange) handleGetBestAsk(c echo.Context) error {
	market := Market(c.Param("market"))
	ob := ex.orderbooks[market]
	order := Order{}

	if len(ob.Asks()) == 0 {
		return c.JSON(http.StatusOK, order)
	}

	bestLimit := ob.Asks()[0]
	bestOrder := bestLimit.Orders[0]

	order.Price = bestLimit.Price
	order.UserID = bestOrder.UserID

	return c.JSON(http.StatusOK, order)

}

func (ex *Exchange) cancelOrder(c echo.Context) error {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	ob := ex.orderbooks[MarketINN]
	order := ob.Orders[int64(id)]
	ob.CancelOrder(order)

	log.Println("order cancelled id =>", id)

	return c.JSON(200, map[string]any{"msg": fmt.Sprintf("order cancelled id => %d", id)})
}

func (ex *Exchange) handlePlaceMarketOrder(market Market, order *orderbook.Order) ([]orderbook.Match, []*MatchedOrder) {
	ob := ex.orderbooks[market]
	matches := ob.PlaceMarketOrder(order)
	matchedOrders := make([]*MatchedOrder, len(matches))

	isBid := false
	if order.Bid {
		isBid = true
	}

	totalSizeFilled := 0.0
	sumPrice := 0.0
	for i := 0; i < len(matchedOrders); i++ {
		limitUserID := matches[i].Bid.UserID
		id := matches[i].Bid.ID
		if isBid {
			limitUserID = matches[i].Ask.UserID
			id = matches[i].Ask.ID
		}
		matchedOrders[i] = &MatchedOrder{
			UserID: limitUserID,
			ID:     id,
			Size:   matches[i].Sizefilled,
			Price:  matches[i].Price,
		}
		totalSizeFilled += matches[i].Sizefilled
		sumPrice += matches[i].Price
	}

	avgPrice := sumPrice / float64(len(matches))

	logrus.WithFields(logrus.Fields{
		"type":     order.Type(),
		"size":     totalSizeFilled,
		"avgPrice": avgPrice,
	}).Info("filled MARKET order")

	newOrderMap := make(map[string][]*orderbook.Order)

	ex.mu.Lock()
	for userID, orderbookOrders := range ex.Orders {
		for i := 0; i < len(orderbookOrders); i++ {
			// If the order is not filled we place it in the map copy.
			// this means that the size of the order == o
			if !orderbookOrders[i].IsFilled() {
				newOrderMap[userID] = append(newOrderMap[userID], orderbookOrders[i])

			}
		}
	}

	ex.Orders = newOrderMap
	ex.mu.Unlock()

	return matches, matchedOrders
}

func (ex *Exchange) handlePlaceLimitOrder(market Market, price float64, order *orderbook.Order) error {
	ob := ex.orderbooks[market]
	ob.PlaceLimitOrder(price, order)

	// keep track of user orders
	ex.mu.Lock()
	ex.Orders[order.UserID] = append(ex.Orders[order.UserID], order)
	ex.mu.Unlock()

	// logrus.WithFields(logrus.Fields{
	// 	"type":  order.Type(),
	// 	"price": order.Limit.Price,
	// 	"size":  order.Size,
	// }).Info("new LIMIT order")

	return nil
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {
	var placeOrderData PlaceOrderRequest

	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return err
	}

	market := Market(placeOrderData.Market)
	order := orderbook.NewOrder(placeOrderData.Bid, placeOrderData.Size, placeOrderData.UserID)

	//Limit Orders
	if placeOrderData.Type == LimitOrder {
		if err := ex.handlePlaceLimitOrder(market, placeOrderData.Price, order); err != nil {
			return err
		}
	}

	//Market Orders
	if placeOrderData.Type == MarketOrder {
		matches, _ := ex.handlePlaceMarketOrder(market, order)
		if err := ex.handleMatches(matches); err != nil {
			return err
		}

	}
	res := &PlaceOrderResponse{
		OrderID: order.ID,
	}

	return c.JSON(http.StatusOK, res)
}

func (ex *Exchange) handleMatches(matches []orderbook.Match) error {
	for _, match := range matches {
		fromUser, ok := ex.Users[match.Ask.UserID]
		if !ok {
			return fmt.Errorf("user not found: %+v", match.Ask.UserID)
		}

		toUser, ok := ex.Users[match.Bid.UserID]
		if !ok {
			return fmt.Errorf("user not found: %+v", match.Bid.UserID)
		}

		toAddress := crypto.PubkeyToAddress(toUser.PrivateKey.PublicKey)

		// for exchange fees
		// exchangePubKey := ex.PrivateKey.Public()
		// publicKeyECDSA, ok := exchangePubKey.(*ecdsa.PublicKey)
		// if !ok {
		// 	return fmt.Errorf("error casting public key to ECDSA")
		// }

		amount := big.NewInt(int64(match.Sizefilled))

		securitiesTransfer(ex.Client, fromUser.PrivateKey, toAddress, amount)

	}
	return nil
}

func securitiesTransfer(client *ethclient.Client, fromPrivKey *ecdsa.PrivateKey, toAddress common.Address, amount *big.Int) error {
	ctx := context.Background()
	publicKey := fromPrivKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatal(err)
	}

	tx := types.NewTransaction(nonce, toAddress, amount, gasLimit, gasPrice, nil)

	chainID := big.NewInt(1337)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), fromPrivKey)
	if err != nil {
		return err
	}

	return client.SendTransaction(ctx, signedTx)

}

func (ex *Exchange) registerUser(pk string, userId string) {
	user := NewUser(pk, userId)
	ex.Users[userId] = user

	logrus.WithFields(logrus.Fields{
		"id": userId,
	}).Info("new exchange user")
}
