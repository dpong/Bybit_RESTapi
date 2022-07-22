package bybitapi

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	Contract   = "CONTRACT"
	Spot       = "SPOT"
	Investment = "INVESTMENT"
	USDT       = "USDT"
)

type CreateInternalTransferResponse struct {
	RetCode int    `json:"ret_code"`
	RetMsg  string `json:"ret_msg"`
	ExtCode string `json:"ext_code"`
	Result  struct {
		TransferID string `json:"transfer_id"`
	} `json:"result"`
	ExtInfo          interface{} `json:"ext_info"`
	TimeNow          int64       `json:"time_now"`
	RateLimitStatus  int         `json:"rate_limit_status"`
	RateLimitResetMs int64       `json:"rate_limit_reset_ms"`
	RateLimit        int         `json:"rate_limit"`
}

func (p *Client) CreateInternalTransfer(coin, from, to string, amount decimal.Decimal) (result *CreateInternalTransferResponse, err error) {
	params := make(map[string]string)
	id := uuid.New()
	params["transfer_id"] = id.String()
	params["coin"] = coin
	params["amount"] = amount.String()
	params["from_account_type"] = from
	params["to_account_type"] = to
	q := url.Values{}
	if params != nil {
		for k, v := range params {
			q.Add(k, v)
		}
	}
	timestamp := time.Now().UnixNano() / 1e6
	q.Add("api_key", p.key)
	q.Add("timestamp", strconv.Itoa(int(timestamp)))
	par := q.Encode()
	signature := p.getSigned(par)
	params["api_key"] = p.key
	params["timestamp"] = strconv.Itoa(int(timestamp))
	params["sign"] = signature
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	res, err := p.sendRequest(ProductPerp, http.MethodPost, "/asset/v1/private/transfer", body, nil, true)
	if err != nil {
		return nil, err
	}
	err = decode(res, &result)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("response is nil")
	}
	if result.RetCode != 0 {
		message := fmt.Sprintf("ret_code=%d, ret_msg=%s, ext_code=%s, ext_info=%s", result.RetCode, result.RetMsg, result.ExtCode, result.ExtInfo)
		return nil, errors.New(message)
	}
	return result, nil
}
