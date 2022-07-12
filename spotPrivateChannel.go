package bybitapi

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type spotPrivateChannelBranch struct {
	cancel     *context.CancelFunc
	key        string
	secret     string
	subaccount string
	product    string
	tradeSets  tradeDataMap
	logger     *logrus.Logger
}

type UserTradeData struct {
	Symbol    string
	Side      string
	Oid       string
	OrderType string
	IsMaker   bool
	Price     decimal.Decimal
	Qty       decimal.Decimal
	Fee       decimal.Decimal
	FeeAsset  string
	TimeStamp time.Time
}

type tradeDataMap struct {
	mux sync.RWMutex
	set map[string][]UserTradeData
}

func (c *Client) CloseSpotPrivateChannel() {
	(*c.spotPrivateChannel.cancel)()
}

func (c *Client) InitSpotPrivateChannel(logger *log.Logger) {
	c.spotPrivateChannelStream(logger)
}

// err is no trade set
func (c *Client) ReadSpotUserTradeWithSymbol(symbol string) ([]UserTradeData, error) {
	c.spotPrivateChannel.tradeSets.mux.Lock()
	defer c.spotPrivateChannel.tradeSets.mux.Unlock()
	uSymbol := strings.ToUpper(symbol)
	var result []UserTradeData
	if data, ok := c.spotPrivateChannel.tradeSets.set[uSymbol]; !ok {
		return data, errors.New("no trade set can be requested")
	} else {
		new := []UserTradeData{}
		result = data
		c.spotPrivateChannel.tradeSets.set[uSymbol] = new
	}
	return result, nil
}

// err is no trade
// mix up with multiple symbol's trade data
func (c *Client) ReadSpotUserTrade() ([]UserTradeData, error) {
	c.spotPrivateChannel.tradeSets.mux.Lock()
	defer c.spotPrivateChannel.tradeSets.mux.Unlock()
	var result []UserTradeData
	for key, item := range c.spotPrivateChannel.tradeSets.set {
		// each symbol
		result = append(result, item...)
		// earse old data
		new := []UserTradeData{}
		c.spotPrivateChannel.tradeSets.set[key] = new
	}
	if len(result) == 0 {
		return result, errors.New("no trade data")
	}
	return result, nil
}

func (c *Client) spotPrivateChannelStream(logger *logrus.Logger) {
	o := new(spotPrivateChannelBranch)
	ctx, cancel := context.WithCancel(context.Background())
	o.cancel = &cancel
	o.key = c.key
	o.secret = c.secret
	o.subaccount = c.subaccount
	o.product = ProductSpot
	o.tradeSets.set = make(map[string][]UserTradeData, 5)
	o.logger = logger
	go o.maintainSession(ctx)
	c.spotPrivateChannel = o
}

func (u *spotPrivateChannelBranch) insertTrade(input *UserTradeData) {
	u.tradeSets.mux.Lock()
	defer u.tradeSets.mux.Unlock()
	if _, ok := u.tradeSets.set[input.Symbol]; !ok {
		// not in the map yet
		data := []UserTradeData{*input}
		u.tradeSets.set[input.Symbol] = data
	} else {
		// already in the map
		data := u.tradeSets.set[input.Symbol]
		data = append(data, *input)
		u.tradeSets.set[input.Symbol] = data
	}
}

func (o *spotPrivateChannelBranch) maintainSession(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := o.maintain(ctx); err == nil {
				return
			} else {
				o.logger.Warningf("reconnect Bybit private channel with err: %s\n", err.Error())
			}
		}
	}
}

func (o *spotPrivateChannelBranch) maintain(ctx context.Context) error {
	var duration time.Duration = 300
	var w ws
	innerErr := make(chan error, 1)
	url := "wss://stream.bybit.com/spot/ws"
	// wait 5 second, if the hand shake fail, will terminate the dail
	dailCtx, _ := context.WithDeadline(ctx, time.Now().Add(time.Second*5))
	conn, _, err := websocket.DefaultDialer.DialContext(dailCtx, url, nil)
	if err != nil {
		return err
	}
	w.conn = conn
	defer w.conn.Close()
	if err := w.getSpotAuth(o.key, o.secret); err != nil {
		return err
	}
	if err := w.conn.SetReadDeadline(time.Now().Add(time.Second * duration)); err != nil {
		return err
	}
	w.conn.SetPingHandler(nil)
	go func() {
		PingManaging := time.NewTicker(time.Second * 30)
		defer PingManaging.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-innerErr:
				return
			case <-PingManaging.C:
				if err := w.sendPingPong(Spot); err != nil {
					w.conn.SetReadDeadline(time.Now().Add(time.Millisecond * 5))
					return
				}
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			_, msg, err := w.conn.ReadMessage()
			if err != nil {
				innerErr <- errors.New("restart")
				return err
			}
			res, err1 := o.decodingInterface(&msg)
			if err1 != nil {
				innerErr <- errors.New("restart")
				return err1
			}
			err2 := o.handleBybitPrivateChannel(o.product, &res)
			if err2 != nil {
				innerErr <- errors.New("restart")
				return err2
			}
			if err := w.conn.SetReadDeadline(time.Now().Add(time.Second * duration)); err != nil {
				innerErr <- errors.New("restart")
				return err
			}
		} // end select
	} // end for
}

// official github
func (w *ws) getSpotAuth(key, secret string) error {
	//generate signature
	expires := fmt.Sprintf("%v", time.Now().Unix()) + "1000"
	h := hmac.New(sha256.New, []byte(secret))
	_val := "GET/realtime" + expires
	io.WriteString(h, _val)
	sign := fmt.Sprintf("%x", h.Sum(nil))
	//auth
	args := []string{key, expires, sign}
	param := make(map[string]interface{})
	param["op"] = "auth"
	param["args"] = args
	req, err := json.Marshal(param)
	if err != nil {
		return err
	}
	// sending
	if err := w.conn.WriteMessage(websocket.TextMessage, req); err != nil {
		return err
	}
	return nil
}

func (o *spotPrivateChannelBranch) decodingInterface(message *[]byte) (res interface{}, err error) {
	err = json.Unmarshal(*message, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (o *spotPrivateChannelBranch) handleBybitPrivateChannel(product string, res *interface{}) error {
	switch message := (*res).(type) {
	case map[string]interface{}:
		if _, ok := message["topic"].(string); !ok {
			if req, ok := message["request"]; ok {
				switch request := req.(type) {
				case string:
					switch {
					case request == "ping":
						if !message["success"].(bool) {
							return errors.New("error on Bybit socket pingpong.")
						}
					}
				case map[string]interface{}:
					switch request["op"].(string) {
					case "auth":
						if !message["success"].(bool) {
							return errors.New("error on Bybit private channel auth.")
						}
					case "subscribe":
						chs := request["args"].([]interface{})
						for _, ch := range chs {
							log.Printf("Subscribed to Bybit %s %s\n", o.product, ch.(string))
						}
					}
				}
			}
			if auth, ok := message["auth"].(string); ok {
				switch auth {
				case "success":
					log.Printf("Subscribed to Bybit %s private channel\n", o.product)
				default:
					return errors.New("fail to Bybit private channel auth.")
				}
			}
		}
	case []interface{}:
		data := message[0].(map[string]interface{})
		if e, ok := data["e"].(string); ok {
			switch e {
			case "executionReport":
				// order handling
				o.handleReport(data)
			case "outboundAccountInfo":
				// balance handling
			case "ticketInfo":
				// pass
			default:
				// pass
			}
		}
	default:

	}
	return nil
}

func (o *spotPrivateChannelBranch) handleReport(data map[string]interface{}) {
	status, ok := data["X"].(string)
	if !ok {
		return
	}
	switch {
	case status == Filled || status == PartialFilled:
		trade := new(UserTradeData)
		if ts, ok := data["E"].(string); ok {
			tsDec, _ := decimal.NewFromString(ts)
			timeStamp := time.UnixMicro(int64(tsDec.InexactFloat64() * 1000))
			trade.TimeStamp = timeStamp
		}
		if s, ok := data["s"].(string); ok {
			trade.Symbol = s
		}
		if q, ok := data["l"].(string); ok {
			qDec, _ := decimal.NewFromString(q)
			trade.Qty = qDec
		}
		if p, ok := data["L"].(string); ok {
			pDec, _ := decimal.NewFromString(p)
			trade.Price = pDec
		}
		if o, ok := data["i"].(string); ok {
			trade.Oid = o
		}
		if S, ok := data["S"].(string); ok {
			if strings.EqualFold(S, "buy") {
				trade.Side = UserTradeBuy
			} else {
				trade.Side = UserTradeSell
			}
		}
		if f, ok := data["n"].(string); ok {
			fDec, _ := decimal.NewFromString(f)
			trade.Fee = fDec
		}
		if m, ok := data["m"].(bool); ok {
			trade.IsMaker = m
		}
		if N, ok := data["N"].(string); ok {
			trade.FeeAsset = N
		}
		if o, ok := data["o"].(string); ok {
			trade.OrderType = o
		}
		// insert
		o.insertTrade(trade)
	default:
		// later
	}

}
