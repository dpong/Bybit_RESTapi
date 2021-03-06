package bybitapi

import (
	"errors"
	"fmt"
	"net/http"
)

type GetSpotWalletBalanceResponse struct {
	RetCode int         `json:"ret_code"`
	RetMsg  string      `json:"ret_msg"`
	ExtCode interface{} `json:"ext_code"`
	ExtInfo interface{} `json:"ext_info"`
	Result  struct {
		Balances []struct {
			Coin     string `json:"coin"`
			Coinid   string `json:"coinId"`
			Coinname string `json:"coinName"`
			Total    string `json:"total"`
			Free     string `json:"free"`
			Locked   string `json:"locked"`
		} `json:"balances"`
	} `json:"result"`
}

func (p *Client) GetSpotWalletBalance() (result *GetSpotWalletBalanceResponse, err error) {
	res, err := p.sendRequest("spot", http.MethodGet, "/spot/v1/account", nil, nil, true)
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
		message := fmt.Sprintf("ret_code=%d, ret_msg=%s, ext_code=%d, ext_info=%s", result.RetCode, result.RetMsg, result.ExtCode, result.ExtInfo)
		return nil, errors.New(message)
	}
	return result, nil
}

type GetSpotServerTimeResponse struct {
	RetCode int         `json:"ret_code"`
	RetMsg  string      `json:"ret_msg"`
	ExtCode interface{} `json:"ext_code"`
	ExtInfo interface{} `json:"ext_info"`
	Result  struct {
		Servertime int64 `json:"serverTime"`
	} `json:"result"`
}

func (p *Client) GetSpotServerTime() (result *GetSpotServerTimeResponse, err error) {
	res, err := p.sendRequest("spot", http.MethodGet, "/spot/v1/time", nil, nil, false)
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
		message := fmt.Sprintf("ret_code=%d, ret_msg=%s, ext_code=%d, ext_info=%s", result.RetCode, result.RetMsg, result.ExtCode, result.ExtInfo)
		return nil, errors.New(message)
	}
	return result, nil
}

type SpotsInfoResponse struct {
	RetCode int         `json:"ret_code"`
	RetMsg  string      `json:"ret_msg"`
	ExtCode interface{} `json:"ext_code"`
	ExtInfo interface{} `json:"ext_info"`
	Result  []struct {
		Name              string `json:"name"`
		Alias             string `json:"alias"`
		Basecurrency      string `json:"baseCurrency"`
		Quotecurrency     string `json:"quoteCurrency"`
		Baseprecision     string `json:"basePrecision"`
		Quoteprecision    string `json:"quotePrecision"`
		Mintradequantity  string `json:"minTradeQuantity"`
		Mintradeamount    string `json:"minTradeAmount"`
		Minpriceprecision string `json:"minPricePrecision"`
		Maxtradequantity  string `json:"maxTradeQuantity"`
		Maxtradeamount    string `json:"maxTradeAmount"`
		Category          int    `json:"category"`
	} `json:"result"`
}

func (p *Client) SpotsInfo() (resp *SpotsInfoResponse, err error) {
	res, err := p.sendRequest("spot", http.MethodGet, "/spot/v1/symbols", nil, nil, false)
	if err != nil {
		return nil, err
	}
	// in Close()
	err = decode(res, &resp)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("response is nil")
	}
	if resp.RetCode != 0 {
		message := fmt.Sprintf("ret_code=%d, ret_msg=%s, ext_code=%d, ext_info=%s", resp.ExtCode, resp.RetMsg, resp.ExtCode, resp.ExtInfo)
		return nil, errors.New(message)
	}
	return resp, nil
}
