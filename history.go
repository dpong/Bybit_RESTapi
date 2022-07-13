package bybitapi

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type rawSpotHistoryKlineResponse struct {
	RetCode int             `json:"ret_code"`
	RetMsg  string          `json:"ret_msg"`
	ExtCode interface{}     `json:"ext_code"`
	ExtInfo interface{}     `json:"ext_info"`
	Result  [][]interface{} `json:"result"`
}

type SpotHistoryKlineResponse struct {
	RetCode int
	RetMsg  string
	ExtCode interface{}
	ExtInfo interface{}
	Data    []KlineData
}

type KlineData struct {
	Open   decimal.Decimal
	High   decimal.Decimal
	Low    decimal.Decimal
	Close  decimal.Decimal
	Volume decimal.Decimal
	Time   time.Time
}

// opts for inteval: 1m, 3m, 5m, 15m, 30m, 1h, 2h, 4h, 6h, 12h, 1d, 1w, 1M
func (p *Client) SpotHistoryKline(symbol, interval string, start, end time.Time) (result *SpotHistoryKlineResponse, err error) {
	params := make(map[string]string)
	params["symbol"] = strings.ToUpper(symbol)
	params["interval"] = interval
	params["startTime"] = fmt.Sprintf("%v", start.UnixMilli())
	params["endTime"] = fmt.Sprintf("%v", end.UnixMilli())
	res, err := p.sendRequest("spot", http.MethodGet, "/spot/quote/v1/kline", nil, &params, false)
	if err != nil {
		return nil, err
	}
	// in Close()
	raw := new(rawSpotHistoryKlineResponse)
	err = decode(res, raw)
	if err != nil {
		return nil, err
	}
	result = new(SpotHistoryKlineResponse)
	result.RetCode = raw.RetCode
	result.RetMsg = raw.RetMsg
	result.ExtCode = raw.ExtCode
	result.ExtInfo = raw.ExtInfo
	var dataList []KlineData
	for _, item := range raw.Result {
		ts := time.UnixMilli(int64(item[0].(float64)))
		open, _ := decimal.NewFromString(item[1].(string))
		high, _ := decimal.NewFromString(item[2].(string))
		low, _ := decimal.NewFromString(item[3].(string))
		close, _ := decimal.NewFromString(item[4].(string))
		vol, _ := decimal.NewFromString(item[5].(string))
		data := KlineData{
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: vol,
			Time:   ts,
		}
		dataList = append(dataList, data)
	}
	result.Data = dataList
	return result, nil
}
