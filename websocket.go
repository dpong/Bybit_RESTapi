package bybitapi

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type ws struct {
	logger *log.Logger
	conn   *websocket.Conn
}

type bybitPing struct {
	Op   string `json:"op,omitempty"`
	Ping int64  `json:"ping,omitempty"`
}

func decodingMap(message *[]byte) (res map[string]interface{}, err error) {
	err = json.Unmarshal(*message, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (w *ws) sendBybitSubscribeMessage(product, channel string, symbols []string) error {
	param := make(map[string]interface{})
	switch product {
	case "perp":
		param["op"] = "subscribe"
		var args []string
		var buffer bytes.Buffer
		for _, symbol := range symbols {
			buffer.WriteString(channel)
			buffer.WriteString(".")
			buffer.WriteString(symbol)
			args = append(args, buffer.String())
			buffer.Reset()
		}
		param["args"] = args
		req, err := json.Marshal(param)
		if err != nil {
			return err
		}
		if err := w.conn.WriteMessage(websocket.TextMessage, req); err != nil {
			return err
		}
	case "spot":
		for _, symbol := range symbols {
			param["event"] = "sub"
			param["topic"] = channel
			inside := make(map[string]interface{})
			inside["symbol"] = symbol
			inside["binary"] = false
			param["params"] = inside
			req, err := json.Marshal(param)
			if err != nil {
				return err
			}
			if err := w.conn.WriteMessage(websocket.TextMessage, req); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *ws) sendPingPong(product string) error {
	mm := bybitPing{}
	switch product {
	case "perp":
		mm.Op = "ping"
	case "spot":
		mm.Ping = time.Now().UnixNano() / 1e6
	}
	message, err := json.Marshal(mm)
	if err != nil {
		return err
	}
	if err := w.conn.WriteMessage(websocket.TextMessage, message); err != nil {
		return err
	}
	return nil
}

func handleBybitSocketData(symobl string, res *map[string]interface{}, mainCh *chan map[string]interface{}) error {
	channel, ok := (*res)["topic"].(string)
	switch ok {
	case true:
		switch channel {
		case "bookTicker":
			if ticker, ok := (*res)["params"].(string); ok {
				if ticker != symobl {
					return nil
				}
			}
			data, ok2 := (*res)["data"].(map[string]interface{})
			if ok2 {
				*mainCh <- data
			}
		case "trade":
			if ticker, ok := (*res)["params"].(string); ok {
				if ticker != symobl {
					return nil
				}
			}
			data, ok2 := (*res)["data"].(map[string]interface{})
			if ok2 {
				*mainCh <- data
			}
		default:
			//
		}
	case false:
		request, check := (*res)["request"].(string)
		switch check {
		case true:
			if request == "ping" {
				if !(*res)["success"].(bool) {
					return errors.New("error on Bybit socket pingpong.")
				}
			}
		}
	}
	return nil
}

func bybitSocket(
	ctx context.Context,
	product, symbol string,
	channel string,
	logger *log.Logger,
	mainCh *chan map[string]interface{},
	reCh *chan error,
) error {
	var w ws
	var duration time.Duration = 45
	w.logger = logger
	innerErr := make(chan error, 1)
	symbol = strings.ToUpper(symbol)
	var url string
	switch product {
	case "perp":
		url = "wss://stream.bybit.com/realtime_public"
	case "spot":
		url = "wss://stream.bybit.com/spot/quote/ws/v2"
	}
	// wait 5 second, if the hand shake fail, will terminate the dail
	dailCtx, _ := context.WithDeadline(ctx, time.Now().Add(time.Second*5))
	conn, _, err := websocket.DefaultDialer.DialContext(dailCtx, url, nil)
	if err != nil {
		return err
	}
	logger.Infof("Bybit %s %s stream connected.\n", symbol, channel)
	w.conn = conn
	defer conn.Close()
	if err := w.sendBybitSubscribeMessage(product, channel, []string{symbol}); err != nil {
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
				if err := w.sendPingPong(product); err != nil {
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
		case err := <-(*reCh):
			innerErr <- errors.New("restart")
			return err
		default:
			_, buf, err := conn.ReadMessage()
			if err != nil {
				innerErr <- errors.New("restart")
				return err
			}
			res, err1 := decodingMap(&buf)
			if err1 != nil {
				innerErr <- errors.New("restart")
				return err1
			}

			err2 := handleBybitSocketData(symbol, &res, mainCh)
			if err2 != nil {
				innerErr <- errors.New("restart")
				return err2
			}
			if err := w.conn.SetReadDeadline(time.Now().Add(time.Second * duration)); err != nil {
				return err
			}
		}
	}
}
