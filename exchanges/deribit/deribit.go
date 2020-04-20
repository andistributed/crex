package deribit

import (
	"errors"
	. "github.com/coinrust/crex"
	"github.com/frankrap/deribit-api"
	"github.com/frankrap/deribit-api/models"
	"time"
)

// Deribit the deribit exchange
type Deribit struct {
	client *deribit.Client
}

func (b *Deribit) GetName() (name string) {
	return "deribit"
}

func (b *Deribit) GetBalance(currency string) (result Balance, err error) {
	params := &models.GetAccountSummaryParams{
		Currency: currency,
		Extended: false,
	}
	var ret models.AccountSummary
	ret, err = b.client.GetAccountSummary(params)
	if err != nil {
		return
	}
	result.Equity = ret.Equity
	result.Available = ret.Balance
	result.RealizedPnl = ret.SessionRpl
	result.UnrealisedPnl = ret.SessionUpl
	return
}

func (b *Deribit) GetOrderBook(symbol string, depth int) (result OrderBook, err error) {
	params := &models.GetOrderBookParams{
		InstrumentName: symbol,
		Depth:          depth,
	}
	var ret models.GetOrderBookResponse
	ret, err = b.client.GetOrderBook(params)
	if err != nil {
		return
	}
	for _, v := range ret.Asks {
		result.Asks = append(result.Asks, Item{
			Price:  v[0],
			Amount: v[1],
		})
	}
	for _, v := range ret.Bids {
		result.Bids = append(result.Bids, Item{
			Price:  v[0],
			Amount: v[1],
		})
	}
	result.Time = time.Unix(0, ret.Timestamp*int64(time.Millisecond)) // 1581819533335
	return
}

func (b *Deribit) GetRecords(symbol string, period string, from int64, end int64, limit int) (records []Record, err error) {
	if end == 0 {
		end = time.Now().Unix()
	}
	params := &models.GetTradingviewChartDataParams{
		InstrumentName: symbol,
		StartTimestamp: from * 1000,
		EndTimestamp:   end * 1000,
		Resolution:     period,
	}
	var resp models.GetTradingviewChartDataResponse
	resp, err = b.client.GetTradingviewChartData(params)
	if err != nil {
		return
	}
	n := len(resp.Ticks)
	for i := 0; i < n; i++ {
		records = append(records, Record{
			Symbol:    symbol,
			Timestamp: time.Unix(0, resp.Ticks[i]*int64(time.Millisecond)),
			Open:      resp.Open[i],
			High:      resp.High[i],
			Low:       resp.Low[i],
			Close:     resp.Close[i],
			Volume:    resp.Volume[i],
		})
	}
	return
}

func (b *Deribit) SetContractType(currencyPair string, contractType string) (err error) {
	return
}

func (b *Deribit) GetContractID() (symbol string, err error) {
	return
}

func (b *Deribit) SetLeverRate(value float64) (err error) {
	return
}

func (b *Deribit) PlaceOrder(symbol string, direction Direction, orderType OrderType, price float64,
	stopPx float64, size float64, postOnly bool, reduceOnly bool, params map[string]interface{}) (result Order, err error) {
	var _orderType string
	var trigger string
	if orderType == OrderTypeLimit {
		_orderType = models.OrderTypeLimit
		stopPx = 0
	} else if orderType == OrderTypeMarket {
		_orderType = models.OrderTypeMarket
		stopPx = 0
	} else if orderType == OrderTypeStopLimit {
		_orderType = models.OrderTypeStopLimit
		trigger = models.TriggerTypeLastPrice
	} else if orderType == OrderTypeStopMarket {
		_orderType = models.OrderTypeStopMarket
		trigger = models.TriggerTypeLastPrice
	}
	if direction == Buy {
		var ret models.BuyResponse
		ret, err = b.client.Buy(&models.BuyParams{
			InstrumentName: symbol,
			Amount:         size,
			Type:           _orderType,
			//Label:          "",
			Price: price,
			//TimeInForce:    "",
			//MaxShow:        nil,
			PostOnly:   postOnly,
			ReduceOnly: reduceOnly,
			StopPrice:  stopPx,
			Trigger:    trigger,
			//Advanced:       "",
		})
		if err != nil {
			return
		}
		result = b.convertOrder(&ret.Order)
	} else if direction == Sell {
		var ret models.SellResponse
		ret, err = b.client.Sell(&models.SellParams{
			InstrumentName: symbol,
			Amount:         size,
			Type:           _orderType,
			//Label:          "",
			Price: price,
			//TimeInForce:    "",
			//MaxShow:        nil,
			PostOnly:   postOnly,
			ReduceOnly: reduceOnly,
			StopPrice:  stopPx,
			Trigger:    trigger,
			//Advanced:       "",
		})
		if err != nil {
			return
		}
		result = b.convertOrder(&ret.Order)
	}
	return
}

func (b *Deribit) GetOpenOrders(symbol string) (result []Order, err error) {
	var ret []models.Order
	ret, err = b.client.GetOpenOrdersByInstrument(&models.GetOpenOrdersByInstrumentParams{
		InstrumentName: symbol,
		//Type:           "",
	})
	if err != nil {
		return
	}
	for _, v := range ret {
		result = append(result, b.convertOrder(&v))
	}
	return
}

func (b *Deribit) GetOrder(symbol string, id string) (result Order, err error) {
	var ret models.Order
	ret, err = b.client.GetOrderState(&models.GetOrderStateParams{
		OrderID: id,
	})
	if err != nil {
		return
	}
	result = b.convertOrder(&ret)
	return
}

func (b *Deribit) CancelOrder(symbol string, id string) (result Order, err error) {
	var order models.Order
	order, err = b.client.Cancel(&models.CancelParams{OrderID: id})
	if err != nil {
		return
	}
	result = b.convertOrder(&order)
	return
}

func (b *Deribit) CancelAllOrders(symbol string) (err error) {
	_, err = b.client.CancelAllByInstrument(&models.CancelAllByInstrumentParams{
		InstrumentName: symbol,
	})
	return
}

func (b *Deribit) AmendOrder(symbol string, id string, price float64, size float64) (result Order, err error) {
	params := &models.EditParams{
		OrderID:   id,
		Amount:    0,
		Price:     0,
		PostOnly:  false,
		Advanced:  "",
		StopPrice: 0,
	}
	if price <= 0 {
		err = errors.New("price is required")
		return
	}
	if size <= 0 {
		err = errors.New("size is required")
		return
	}
	params.Price = price
	params.Amount = size
	var resp models.EditResponse
	resp, err = b.client.Edit(params)
	if err != nil {
		return
	}
	result = b.convertOrder(&resp.Order)
	return
}

func (b *Deribit) GetPositions(symbol string) (result []Position, err error) {
	var ret models.Position
	ret, err = b.client.GetPosition(&models.GetPositionParams{InstrumentName: symbol})
	if err != nil {
		return
	}
	result = []Position{
		{
			Symbol:    symbol,
			OpenTime:  time.Time{},
			OpenPrice: ret.AveragePrice,
			Size:      ret.Size,
			AvgPrice:  ret.AveragePrice,
		},
	}
	return
}

func (b *Deribit) convertOrder(order *models.Order) (result Order) {
	result.ID = order.OrderID
	result.Symbol = order.InstrumentName
	result.Price = order.Price.ToFloat64()
	result.StopPx = order.StopPrice
	result.Size = order.Amount
	result.Direction = b.convertDirection(order.Direction)
	result.Type = b.convertOrderType(order.OrderType)
	result.AvgPrice = order.AveragePrice
	result.FilledAmount = order.FilledAmount
	result.PostOnly = order.PostOnly
	result.ReduceOnly = order.ReduceOnly
	result.Status = b.orderStatus(order)
	return
}

func (b *Deribit) convertDirection(direction string) Direction {
	switch direction {
	case models.DirectionBuy:
		return Buy
	case models.DirectionSell:
		return Sell
	default:
		return Buy
	}
}

func (b *Deribit) convertOrderType(orderType string) OrderType {
	switch orderType {
	case models.OrderTypeLimit:
		return OrderTypeLimit
	case models.OrderTypeMarket:
		return OrderTypeMarket
	case models.OrderTypeStopLimit:
		return OrderTypeStopLimit
	case models.OrderTypeStopMarket:
		return OrderTypeStopMarket
	default:
		return OrderTypeLimit
	}
}

func (b *Deribit) orderStatus(order *models.Order) OrderStatus {
	orderState := order.OrderState
	switch orderState {
	case models.OrderStateOpen:
		if order.FilledAmount > 0 {
			return OrderStatusPartiallyFilled
		}
		return OrderStatusNew
	case models.OrderStateFilled:
		return OrderStatusFilled
	case models.OrderStateRejected:
		return OrderStatusRejected
	case models.OrderStateCancelled:
		return OrderStatusCancelled
	case models.OrderStateUntriggered:
		return OrderStatusUntriggered
	default:
		return OrderStatusCreated
	}
}

func (b *Deribit) SubscribeTrades(market Market, callback func(trades []Trade)) error {
	return ErrNotImplemented
}

func (b *Deribit) SubscribeLevel2Snapshots(market Market, callback func(ob *OrderBook)) error {
	return ErrNotImplemented
}

func (b *Deribit) SubscribeOrders(market Market, callback func(orders []Order)) error {
	return ErrNotImplemented
}

func (b *Deribit) SubscribePositions(market Market, callback func(positions []Position)) error {
	return ErrNotImplemented
}

func (b *Deribit) RunEventLoopOnce() (err error) {
	return
}

func NewDeribit(params *Parameters) *Deribit {
	baseUri := "wss://www.deribit.com/ws/api/v2/"
	if params.Testnet {
		baseUri = "wss://test.deribit.com/ws/api/v2/"
	}
	cfg := &deribit.Configuration{
		Addr:          baseUri,
		ApiKey:        params.AccessKey,
		SecretKey:     params.SecretKey,
		AutoReconnect: true,
	}
	client := deribit.New(cfg)
	return &Deribit{
		client: client,
	}
}