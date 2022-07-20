package bybitapi

import (
	"net/http"
	"strings"

	"github.com/shopspring/decimal"
)

const (
	StatusNew     = "New"
	StatusPartial = "PartiallyFilled"
)

type PerpPlaceOrderResponse struct {
	RetCode          int                  `json:"ret_code"`
	RetMsg           string               `json:"ret_msg"`
	ExtCode          string               `json:"ext_code"`
	ExtInfo          string               `json:"ext_info"`
	Result           PerpPlaceOrderResult `json:"result"`
	TimeNow          string               `json:"time_now"`
	RateLimitStatus  int                  `json:"rate_limit_status"`
	RateLimitResetMs int64                `json:"rate_limit_reset_ms"`
	RateLimit        int                  `json:"rate_limit"`
}

type PerpPlaceOrderResult struct {
	OrderID        string  `json:"order_id"`
	UserID         int     `json:"user_id"`
	Symbol         string  `json:"symbol"`
	Side           string  `json:"side"`
	OrderType      string  `json:"order_type"`
	Price          float64 `json:"price"`
	Qty            float64 `json:"qty"`
	TimeInForce    string  `json:"time_in_force"`
	OrderStatus    string  `json:"order_status"`
	LastExecPrice  float64 `json:"last_exec_price"`
	CumExecQty     float64 `json:"cum_exec_qty"`
	CumExecValue   float64 `json:"cum_exec_value"`
	CumExecFee     float64 `json:"cum_exec_fee"`
	ReduceOnly     bool    `json:"reduce_only"`
	CloseOnTrigger bool    `json:"close_on_trigger"`
	OrderLinkID    string  `json:"order_link_id"`
	CreatedTime    string  `json:"created_time"`
	UpdatedTime    string  `json:"updated_time"`
}

func (p *Client) PerpPlaceOrder(symbol, side, order_type string, price, qty decimal.Decimal, reduce_only bool) (result *PerpPlaceOrderResponse, err error) {
	params := make(map[string]interface{})
	params["symbol"] = strings.ToUpper(symbol)
	params["side"] = side
	switch order_type {
	case Limit:
		fprice, _ := price.Float64()
		params["price"] = fprice
	}
	fqty, _ := qty.Float64()
	params["qty"] = fqty
	params["order_type"] = order_type
	params["time_in_force"] = GTC
	params["close_on_trigger"] = false
	params["reduce_only"] = reduce_only
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	res, err := p.sendRequest(ProductPerp, http.MethodPost, "/private/linear/order/create", body, nil, true)
	if err != nil {
		return nil, err
	}
	err = decode(res, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type PerpGetOrderResponse struct {
	RetCode int    `json:"ret_code"`
	RetMsg  string `json:"ret_msg"`
	ExtCode string `json:"ext_code"`
	ExtInfo string `json:"ext_info"`
	Result  struct {
		OrderID        string  `json:"order_id"`
		UserID         int     `json:"user_id"`
		Symbol         string  `json:"symbol"`
		Side           string  `json:"side"`
		OrderType      string  `json:"order_type"`
		Price          float64 `json:"price"`
		Qty            float64 `json:"qty"`
		TimeInForce    string  `json:"time_in_force"`
		OrderStatus    string  `json:"order_status"`
		LastExecPrice  float64 `json:"last_exec_price"`
		CumExecQty     float64 `json:"cum_exec_qty"`
		CumExecValue   float64 `json:"cum_exec_value"`
		CumExecFee     float64 `json:"cum_exec_fee"`
		OrderLinkID    string  `json:"order_link_id"`
		ReduceOnly     bool    `json:"reduce_only"`
		CloseOnTrigger bool    `json:"close_on_trigger"`
		CreatedTime    string  `json:"created_time"`
		UpdatedTime    string  `json:"updated_time"`
	} `json:"result"`
	TimeNow          string `json:"time_now"`
	RateLimitStatus  int    `json:"rate_limit_status"`
	RateLimitResetMs int64  `json:"rate_limit_reset_ms"`
	RateLimit        int    `json:"rate_limit"`
}

func (p *Client) PerpGetOrder(symbol, oid string) (result *PerpGetOrderResponse, err error) {
	params := make(map[string]string)
	params["symbol"] = strings.ToUpper(symbol)
	if oid != "" {
		params["order_id"] = oid
	}
	res, err := p.sendRequest(ProductPerp, http.MethodGet, "/private/linear/order/search", nil, &params, true)
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

type PerpCancelOrderResponse struct {
	RetCode int    `json:"ret_code"`
	RetMsg  string `json:"ret_msg"`
	ExtCode string `json:"ext_code"`
	ExtInfo string `json:"ext_info"`
	Result  struct {
		OrderID string `json:"order_id"`
	} `json:"result"`
	TimeNow          string `json:"time_now"`
	RateLimitStatus  int    `json:"rate_limit_status"`
	RateLimitResetMs int64  `json:"rate_limit_reset_ms"`
	RateLimit        int    `json:"rate_limit"`
}

func (p *Client) PerpCancelOrder(symbol, oid string) (result *PerpGetOrderResponse, err error) {
	params := make(map[string]interface{})
	params["symbol"] = strings.ToUpper(symbol)
	params["order_id"] = oid
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	res, err := p.sendRequest(ProductPerp, http.MethodPost, "/private/linear/order/cancel", body, nil, true)
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

type PerpReplaceOrderResponse struct {
	RetCode int    `json:"ret_code"`
	RetMsg  string `json:"ret_msg"`
	ExtCode string `json:"ext_code"`
	Result  struct {
		OrderID string `json:"order_id"`
	} `json:"result"`
	TimeNow          string `json:"time_now"`
	RateLimitStatus  int    `json:"rate_limit_status"`
	RateLimitResetMs int64  `json:"rate_limit_reset_ms"`
	RateLimit        int    `json:"rate_limit"`
}

// can replace price and qty, if dont't want to replace any of them, pass 0
func (p *Client) PerpReplaceOrder(symbol, oid string, price, qty decimal.Decimal) (result *PerpReplaceOrderResponse, err error) {
	params := make(map[string]interface{})
	params["symbol"] = strings.ToUpper(symbol)
	params["order_id"] = oid
	if !price.IsZero() {
		fprice, _ := price.Float64()
		params["p_r_price"] = fprice
	}
	if !price.IsZero() {
		fqty, _ := qty.Float64()
		params["p_r_qty"] = fqty
	}
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	res, err := p.sendRequest(ProductPerp, http.MethodPost, "/private/linear/order/create", body, nil, true)
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

type PerpGetAllOpenOrdersResponse struct {
	RetCode int    `json:"ret_code"`
	RetMsg  string `json:"ret_msg"`
	ExtCode string `json:"ext_code"`
	Result  struct {
		CurrentPage int `json:"current_page"`
		LastPage    int `json:"last_page"`
		Data        []struct {
			OrderID        string  `json:"order_id"`
			UserID         int     `json:"user_id"`
			Symbol         string  `json:"symbol"`
			Side           string  `json:"side"`
			OrderType      string  `json:"order_type"`
			Price          float64 `json:"price"`
			Qty            float64 `json:"qty"`
			TimeInForce    string  `json:"time_in_force"`
			OrderStatus    string  `json:"order_status"`
			LastExecPrice  float64 `json:"last_exec_price"`
			CumExecQty     float64 `json:"cum_exec_qty"`
			CumExecValue   float64 `json:"cum_exec_value"`
			CumExecFee     float64 `json:"cum_exec_fee"`
			OrderLinkID    string  `json:"order_link_id"`
			ReduceOnly     bool    `json:"reduce_only"`
			CloseOnTrigger bool    `json:"close_on_trigger"`
			CreatedTime    string  `json:"created_time"`
			UpdatedTime    string  `json:"updated_time"`
		} `json:"data"`
	} `json:"result"`
	ExtInfo          interface{} `json:"ext_info"`
	TimeNow          string      `json:"time_now"`
	RateLimitStatus  int         `json:"rate_limit_status"`
	RateLimitResetMs int64       `json:"rate_limit_reset_ms"`
	RateLimit        int         `json:"rate_limit"`
}

// New / PartiallyFilled
func (p *Client) PerpGetAllOpenOrders(symbol string) (result *PerpGetAllOpenOrdersResponse, err error) {
	params := make(map[string]string)
	params["symbol"] = strings.ToUpper(symbol)
	res, err := p.sendRequest(ProductPerp, http.MethodGet, "/private/linear/order/list", nil, &params, true)
	if err != nil {
		return nil, err
	}
	err = decode(res, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type PerpCancelAllOrdersResponse struct {
	RetCode          int      `json:"ret_code"`
	RetMsg           string   `json:"ret_msg"`
	ExtCode          string   `json:"ext_code"`
	ExtInfo          string   `json:"ext_info"`
	Result           []string `json:"result"`
	TimeNow          string   `json:"time_now"`
	RateLimitStatus  int      `json:"rate_limit_status"`
	RateLimitResetMs int64    `json:"rate_limit_reset_ms"`
	RateLimit        int      `json:"rate_limit"`
}

// this method will consume 10 requests, be careful
func (p *Client) PerpCancelAllOrders(symbol string) (result *PerpCancelAllOrdersResponse, err error) {
	params := make(map[string]interface{})
	params["symbol"] = strings.ToUpper(symbol)
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	res, err := p.sendRequest(ProductPerp, http.MethodPost, "/private/linear/order/cancel-all", body, nil, true)
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
