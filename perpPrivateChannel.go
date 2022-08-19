package bybitapi

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type perpPrivateChannelBranch struct {
	cancel     *context.CancelFunc
	key        string
	secret     string
	subaccount string
	product    string
	tradeSets  tradeDataMap
	logger     *logrus.Logger
}

func (c *Client) ClosePerpPrivateChannel() {
	(*c.spotPrivateChannel.cancel)()
}

// err is no trade set
func (c *Client) ReadPerpUserTradeWithSymbol(symbol string) ([]UserTradeData, error) {
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
func (c *Client) ReadPerpUserTrade() ([]UserTradeData, error) {
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

func (c *Client) InitPerpPrivateChannel(logger *log.Logger) {
	c.perpPrivateChannelStream(logger)
}

// internal

func (u *perpPrivateChannelBranch) insertTrade(input *UserTradeData) {
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

func (c *Client) perpPrivateChannelStream(logger *logrus.Logger) {
	o := new(perpPrivateChannelBranch)
	ctx, cancel := context.WithCancel(context.Background())
	o.cancel = &cancel
	o.key = c.key
	o.secret = c.secret
	o.subaccount = c.subaccount
	o.product = ProductSpot
	o.tradeSets.set = make(map[string][]UserTradeData, 5)
	o.logger = logger
	go o.maintainSession(ctx)
	c.perpPrivateChannel = o
}

func (o *perpPrivateChannelBranch) maintainSession(ctx context.Context) {
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

func (o *perpPrivateChannelBranch) maintain(ctx context.Context) error {
	var duration time.Duration = 300
	var w ws
	innerErr := make(chan error, 1)
	url := "wss://stream.bybit.com/realtime_private"
	// wait 5 second, if the hand shake fail, will terminate the dail
	dailCtx, _ := context.WithDeadline(ctx, time.Now().Add(time.Second*5))
	conn, _, err := websocket.DefaultDialer.DialContext(dailCtx, url, nil)
	if err != nil {
		return err
	}
	w.conn = conn
	defer w.conn.Close()
	if err := w.getAuth(o.key, o.secret); err != nil {
		return err
	}
	if err := w.conn.SetReadDeadline(time.Now().Add(time.Second * duration)); err != nil {
		return err
	}
	w.conn.SetPingHandler(nil)
	go func() {
		PingManaging := time.NewTicker(time.Second * 15)
		defer PingManaging.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-innerErr:
				return
			case <-PingManaging.C:
				if err := w.sendPingPong(ProductSpot); err != nil {
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
			err2 := o.handleBybitPrivateChannel(o.product, &res, &w)
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

func (o *perpPrivateChannelBranch) decodingInterface(message *[]byte) (res interface{}, err error) {
	err = json.Unmarshal(*message, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (o *perpPrivateChannelBranch) handleBybitPrivateChannel(product string, res *interface{}, w *ws) error {
	switch message := (*res).(type) {
	case map[string]interface{}:
		if channel, ok := message["topic"].(string); ok {
			if channel == "execution" {
				datas := message["data"].([]interface{})
				for _, data := range datas {
					o.handleExecution(data.(map[string]interface{}))
				}
			}
		} else {
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
							o.logger.Printf("Subscribed to Bybit %s %s\n", o.product, ch.(string))
						}
					}
				}
			}
			if auth, ok := message["auth"].(string); ok {
				if auth != "success" {
					return errors.New("fail to Bybit private channel auth.")
				}
				o.logger.Printf("Subscribed to Bybit %s private channel\n", o.product)
				// subscribe execution channel
				w.getPerpPrivateSubscribe("execution")
			}
		}
	}
	return nil
}

// "execution"
func (w *ws) getPerpPrivateSubscribe(channel string) error {
	param := make(map[string]interface{})
	param["op"] = "subscribe"
	param["args"] = []string{channel}
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

func (o *perpPrivateChannelBranch) handleExecution(data map[string]interface{}) {
	trade := new(UserTradeData)
	// timestamp
	if timestamp, ok := data["trade_time"].(string); ok {
		layout := "2006-01-02T15:04:05.999999Z"
		st, _ := time.Parse(layout, timestamp)
		trade.TimeStamp = st
	}
	if q, ok := data["exec_qty"].(float64); ok {
		qDec := decimal.NewFromFloat(q)
		trade.Qty = qDec
	}
	if s, ok := data["symbol"].(string); ok {
		trade.Symbol = s
	}
	if p, ok := data["price"].(float64); ok {
		pDec := decimal.NewFromFloat(p)
		trade.Price = pDec
	}
	if o, ok := data["order_id"].(string); ok {
		trade.Oid = o
	}
	if S, ok := data["side"].(string); ok {
		if strings.EqualFold(S, "buy") {
			trade.Side = UserTradeBuy
		} else {
			trade.Side = UserTradeSell
		}
	}
	if f, ok := data["exec_fee"].(float64); ok {
		fDec := decimal.NewFromFloat(f)
		trade.Fee = fDec
		trade.FeeAsset = "USDT"
	}
	if m, ok := data["is_maker"].(bool); ok {
		trade.IsMaker = m
		if m {
			trade.OrderType = "limit"
		} else {
			trade.OrderType = "market"
		}

	}
	o.insertTrade(trade)
}
