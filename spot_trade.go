package bybitapi

import (
	"net/http"
	"strings"

	"github.com/shopspring/decimal"
)

type SpotPlaceOrderResponse struct {
	RetCode int         `json:"ret_code"`
	RetMsg  string      `json:"ret_msg"`
	ExtCode interface{} `json:"ext_code"`
	ExtInfo interface{} `json:"ext_info"`
	Result  struct {
		Accountid    string `json:"accountId"`
		Symbol       string `json:"symbol"`
		Symbolname   string `json:"symbolName"`
		Orderlinkid  string `json:"orderLinkId"`
		Orderid      string `json:"orderId"`
		Transacttime string `json:"transactTime"`
		Price        string `json:"price"`
		Origqty      string `json:"origQty"`
		Executedqty  string `json:"executedQty"`
		Status       string `json:"status"`
		Timeinforce  string `json:"timeInForce"`
		Type         string `json:"type"`
		Side         string `json:"side"`
	} `json:"result"`
}

// ex: "BTCUSDT"
func (p *Client) SpotPlaceOrder(symbol, side, order_type string, price, qty decimal.Decimal) (result *SpotPlaceOrderResponse, err error) {
	params := make(map[string]string)
	params["symbol"] = strings.ToUpper(symbol)
	params["side"] = side
	switch order_type {
	case Limit:
		params["price"] = price.String()
	}
	params["qty"] = qty.String()
	params["type"] = order_type
	params["time_in_force"] = "GTC"
	/*
		body, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
	*/
	res, err := p.sendRequest("spot", http.MethodPost, "/spot/v1/order", nil, &params, true)
	if err != nil {
		return nil, err
	}
	// in Close()
	err = decode(res, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type SpotCancelOrderResponse struct {
	RetCode int         `json:"ret_code"`
	RetMsg  string      `json:"ret_msg"`
	ExtCode interface{} `json:"ext_code"`
	ExtInfo interface{} `json:"ext_info"`
	Result  struct {
		Accountid    string `json:"accountId"`
		Symbol       string `json:"symbol"`
		Orderlinkid  string `json:"orderLinkId"`
		Orderid      string `json:"orderId"`
		Transacttime string `json:"transactTime"`
		Price        string `json:"price"`
		Origqty      string `json:"origQty"`
		Executedqty  string `json:"executedQty"`
		Status       string `json:"status"`
		Timeinforce  string `json:"timeInForce"`
		Type         string `json:"type"`
		Side         string `json:"side"`
	} `json:"result"`
}

func (p *Client) SpotCancelOrder(oid string) (result *SpotCancelOrderResponse, err error) {
	params := make(map[string]interface{})
	params["orderId"] = oid
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	res, err := p.sendRequest("spot", http.MethodDelete, "/spot/v1/order", body, nil, true)
	if err != nil {
		return nil, err
	}
	// in Close()
	err = decode(res, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type SpotCancelAllOrdersResponse struct {
	RetCode int         `json:"ret_code"`
	RetMsg  string      `json:"ret_msg"`
	ExtCode interface{} `json:"ext_code"`
	ExtInfo interface{} `json:"ext_info"`
	Result  struct {
		Success bool `json:"success"`
	} `json:"result"`
}

func (p *Client) SpotCancelAllOrders(symbol string) (result *SpotCancelOrderResponse, err error) {
	params := make(map[string]interface{})
	params["symbol"] = symbol
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	res, err := p.sendRequest("spot", http.MethodDelete, "/spot/order/batch-cancel", body, nil, true)
	if err != nil {
		return nil, err
	}
	// in Close()
	err = decode(res, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type SpotGetOrderResponse struct {
	RetCode int         `json:"ret_code"`
	RetMsg  string      `json:"ret_msg"`
	ExtCode interface{} `json:"ext_code"`
	ExtInfo interface{} `json:"ext_info"`
	Result  struct {
		Accountid           string `json:"accountId"`
		Exchangeid          string `json:"exchangeId"`
		Symbol              string `json:"symbol"`
		Symbolname          string `json:"symbolName"`
		Orderlinkid         string `json:"orderLinkId"`
		Orderid             string `json:"orderId"`
		Price               string `json:"price"`
		Origqty             string `json:"origQty"`
		Executedqty         string `json:"executedQty"`
		Cummulativequoteqty string `json:"cummulativeQuoteQty"`
		Avgprice            string `json:"avgPrice"`
		Status              string `json:"status"`
		Timeinforce         string `json:"timeInForce"`
		Type                string `json:"type"`
		Side                string `json:"side"`
		Stopprice           string `json:"stopPrice"`
		Icebergqty          string `json:"icebergQty"`
		Time                string `json:"time"`
		Updatetime          string `json:"updateTime"`
		Isworking           bool   `json:"isWorking"`
	} `json:"result"`
}

func (p *Client) SpotGetOrder(oid string) (result *SpotGetOrderResponse, err error) {
	params := make(map[string]interface{})
	params["orderId"] = oid
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	res, err := p.sendRequest("spot", http.MethodGet, "/spot/v1/order", body, nil, true)
	if err != nil {
		return nil, err
	}
	// in Close()
	err = decode(res, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
