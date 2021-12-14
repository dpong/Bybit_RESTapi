package bybitapi

import (
	"net/http"

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
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	res, err := p.sendRequest("swap", http.MethodPost, "/asset/v1/private/transfer", body, nil, true)
	if err != nil {
		return nil, err
	}
	err = decode(res, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
