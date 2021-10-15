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

func (p *Client) SpotPlaceOrder(symbol, side, order_type string, price, qty decimal.Decimal) (result *SpotPlaceOrderResponse, err error) {
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
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	res, err := p.sendRequest("spot", http.MethodPost, "/spot/v1/order", body, nil, true)
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
