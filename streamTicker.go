package bybitapi

import (
	"context"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

const NullPrice = "null"

type StreamTickerBranch struct {
	bid    tobBranch
	ask    tobBranch
	cancel *context.CancelFunc
	reCh   chan error
}

type tobBranch struct {
	mux       sync.RWMutex
	price     string
	qty       string
	timeStamp time.Time
}

func (s *StreamTickerBranch) Close() {
	(*s.cancel)()
	s.bid.mux.Lock()
	s.bid.price = NullPrice
	s.bid.mux.Unlock()
	s.ask.mux.Lock()
	s.ask.price = NullPrice
	s.ask.mux.Unlock()
}

func (s *StreamTickerBranch) GetBid() (price, qty string, timeStamp time.Time, ok bool) {
	s.bid.mux.RLock()
	defer s.bid.mux.RUnlock()
	price = s.bid.price
	qty = s.bid.qty
	timeStamp = s.bid.timeStamp
	if price == NullPrice || price == "" {
		return price, qty, timeStamp, false
	}
	return price, qty, timeStamp, true
}

func (s *StreamTickerBranch) GetAsk() (price, qty string, timeStamp time.Time, ok bool) {
	s.ask.mux.RLock()
	defer s.ask.mux.RUnlock()
	price = s.ask.price
	qty = s.ask.qty
	timeStamp = s.ask.timeStamp
	if price == NullPrice || price == "" {
		return price, qty, timeStamp, false
	}
	return price, qty, timeStamp, true
}

func StreamTickerSpot(symbol string, logger *log.Logger) *StreamTickerBranch {
	return streamTicker(ProductSpot, symbol, logger)
}

func StreamTickerPerp(symbol string, logger *log.Logger) *StreamTickerBranch {
	return streamTicker(ProductPerp, symbol, logger)
}

// internal

func streamTicker(product, symbol string, logger *log.Logger) *StreamTickerBranch {
	var s StreamTickerBranch
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = &cancel
	ticker := make(chan map[string]interface{}, 50)
	errCh := make(chan error, 5)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := bybitSocket(ctx, product, symbol, "bookTicker", logger, &ticker, &errCh); err == nil {
					return
				} else {
					logger.Warningf("Reconnect %s ticker stream with err: %s\n", symbol, err.Error())
				}
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := s.maintainStreamTicker(ctx, symbol, &ticker, &errCh); err == nil {
					return
				} else {
					logger.Warningf("Refreshing %s ticker stream with err: %s\n", symbol, err.Error())
				}
			}
		}
	}()
	return &s
}

func (s *StreamTickerBranch) updateBidData(price, qty string, timeStamp time.Time) {
	s.bid.mux.Lock()
	defer s.bid.mux.Unlock()
	s.bid.price = price
	s.bid.qty = qty
	s.bid.timeStamp = timeStamp
}

func (s *StreamTickerBranch) updateAskData(price, qty string, timeStamp time.Time) {
	s.ask.mux.Lock()
	defer s.ask.mux.Unlock()
	s.ask.price = price
	s.ask.qty = qty
	s.ask.timeStamp = timeStamp
}

func (s *StreamTickerBranch) maintainStreamTicker(
	ctx context.Context,
	symbol string,
	ticker *chan map[string]interface{},
	errCh *chan error,
) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case message := <-(*ticker):
			// millisecond level
			rawTs, ok := message["time"].(float64)
			if !ok {
				continue
			}
			ts := time.UnixMicro(int64(rawTs * 1000))
			var bidPrice, askPrice, bidQty, askQty string
			if bid, ok := message["bidPrice"].(string); ok {
				bidDec, _ := decimal.NewFromString(bid)
				bidPrice = bidDec.String()
			} else {
				bidPrice = NullPrice
			}
			if ask, ok := message["askPrice"].(string); ok {
				askDec, _ := decimal.NewFromString(ask)
				askPrice = askDec.String()
			} else {
				askPrice = NullPrice
			}
			if bidqty, ok := message["bidQty"].(string); ok {
				bidQtyDec, _ := decimal.NewFromString(bidqty)
				bidQty = bidQtyDec.String()
			}
			if askqty, ok := message["askQty"].(string); ok {
				askQtyDec, _ := decimal.NewFromString(askqty)
				askQty = askQtyDec.String()
			}
			s.updateBidData(bidPrice, bidQty, ts)
			s.updateAskData(askPrice, askQty, ts)
		}
	}
}
