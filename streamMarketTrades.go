package bybitapi

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type StreamMarketTradesBranch struct {
	cancel       *context.CancelFunc
	product      string
	symbol       string
	tradeChan    chan map[string]interface{}
	tradesBranch struct {
		Trades []PublicTradeData
		sync.Mutex
	}
	logger *logrus.Logger
}

// taker side
type PublicTradeData struct {
	Product string
	Symbol  string
	Side    string
	Price   decimal.Decimal
	Qty     decimal.Decimal
	Time    time.Time
}

func StreamTradeSpot(symbol string, logger *logrus.Logger) *StreamMarketTradesBranch {
	Usymbol := strings.ToUpper(symbol)
	return streamTrade(ProductSpot, Usymbol, logger)
}

// side: Side of the taker in the trade
func (o *StreamMarketTradesBranch) GetTrades() []PublicTradeData {
	o.tradesBranch.Lock()
	defer o.tradesBranch.Unlock()
	trades := o.tradesBranch.Trades
	o.tradesBranch.Trades = []PublicTradeData{}
	return trades
}

func (o *StreamMarketTradesBranch) Close() {
	(*o.cancel)()
	o.tradesBranch.Lock()
	defer o.tradesBranch.Unlock()
	o.tradesBranch.Trades = []PublicTradeData{}
}

// spot only for now
func streamTrade(product, symbol string, logger *logrus.Logger) *StreamMarketTradesBranch {
	o := new(StreamMarketTradesBranch)
	ctx, cancel := context.WithCancel(context.Background())
	o.cancel = &cancel
	o.product = product
	o.symbol = symbol
	o.tradeChan = make(chan map[string]interface{}, 100)
	o.logger = logger
	errCh := make(chan error, 5)
	go o.maintainSession(ctx, &errCh)
	go o.listen(ctx)
	return o
}

func (o *StreamMarketTradesBranch) listen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case trade := <-o.tradeChan:
			data := new(PublicTradeData)
			data.Symbol = o.symbol
			data.Product = o.product

			if ts, ok := trade["t"].(float64); ok {
				timeStamp := time.UnixMicro(int64(ts * 1000))
				data.Time = timeStamp
			}
			if priceStr, ok := trade["p"].(string); ok {
				priceDec, _ := decimal.NewFromString(priceStr)
				data.Price = priceDec
			}
			if qtyStr, ok := trade["q"].(string); ok {
				qtyDec, _ := decimal.NewFromString(qtyStr)
				data.Qty = qtyDec
			}
			if buyTaker, ok := trade["m"].(bool); ok {
				if buyTaker {
					data.Side = "buy"
				} else {
					data.Side = "sell"
				}
			}
			o.appendNewTrade(data)
		}
	}
}

func (o *StreamMarketTradesBranch) appendNewTrade(new *PublicTradeData) {
	o.tradesBranch.Lock()
	defer o.tradesBranch.Unlock()
	o.tradesBranch.Trades = append(o.tradesBranch.Trades, *new)
}

func (o *StreamMarketTradesBranch) maintainSession(ctx context.Context, errCh *chan error) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := bybitSocket(ctx, o.product, o.symbol, "trade", o.logger, &o.tradeChan, errCh); err == nil {
				return
			} else {
				o.logger.Warningf("reconnect FTX %s trade stream with err: %s\n", o.symbol, err.Error())
			}
		}
	}
}
