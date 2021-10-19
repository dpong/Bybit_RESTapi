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

type SwapPlaceOrderResponse struct {
	RetCode          int                  `json:"ret_code"`
	RetMsg           string               `json:"ret_msg"`
	ExtCode          string               `json:"ext_code"`
	ExtInfo          string               `json:"ext_info"`
	Result           SwapPlaceOrderResult `json:"result"`
	TimeNow          string               `json:"time_now"`
	RateLimitStatus  int                  `json:"rate_limit_status"`
	RateLimitResetMs int64                `json:"rate_limit_reset_ms"`
	RateLimit        int                  `json:"rate_limit"`
}

type SwapPlaceOrderResult struct {
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

func (p *Client) SwapPlaceOrder(symbol, side, order_type string, price, qty decimal.Decimal, reduce_only bool) (result *SwapPlaceOrderResponse, err error) {
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
	res, err := p.sendRequest("swap", http.MethodPost, "/private/linear/order/create", body, nil, true)
	if err != nil {
		return nil, err
	}
	err = decode(res, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type SwapGetOrderResponse struct {
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

func (p *Client) SwapGetOrder(symbol, oid string) (result *SwapGetOrderResponse, err error) {
	params := make(map[string]string)
	params["symbol"] = strings.ToUpper(symbol)
	if oid != "" {
		params["order_id"] = oid
	}
	res, err := p.sendRequest("swap", http.MethodGet, "/private/linear/order/search", nil, &params, true)
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

type SwapCancelOrderResponse struct {
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

func (p *Client) SwapCancelOrder(symbol, oid string) (result *SwapGetOrderResponse, err error) {
	params := make(map[string]interface{})
	params["symbol"] = strings.ToUpper(symbol)
	params["order_id"] = oid
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	res, err := p.sendRequest("swap", http.MethodPost, "/private/linear/order/cancel", body, nil, true)
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

type SwapReplaceOrderResponse struct {
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
func (p *Client) SwapReplaceOrder(symbol, oid string, price, qty decimal.Decimal) (result *SwapReplaceOrderResponse, err error) {
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
	res, err := p.sendRequest("swap", http.MethodPost, "/private/linear/order/create", body, nil, true)
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

type SwapGetAllOpenOrdersResponse struct {
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
func (p *Client) SwapGetAllOpenOrders(symbol, status string) (result *SwapGetAllOpenOrdersResponse, err error) {
	params := make(map[string]string)
	params["symbol"] = strings.ToUpper(symbol)
	params["order_status"] = status
	res, err := p.sendRequest("swap", http.MethodGet, "/private/linear/order/list", nil, &params, true)
	if err != nil {
		return nil, err
	}
	err = decode(res, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type SwapCancelAllOrdersResponse struct {
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
func (p *Client) SwapCancelAllOrders(symbol string) (result *SwapCancelAllOrdersResponse, err error) {
	params := make(map[string]interface{})
	params["symbol"] = strings.ToUpper(symbol)
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	res, err := p.sendRequest("swap", http.MethodPost, "/private/linear/order/cancel-all", body, nil, true)
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
