package bybitapi

import (
	"strings"

	"github.com/shopspring/decimal"
)

// [][]string{oid, symbol, product, subaccount, price, qty, side, orderType, fee, filledQty, timestamp, isMaker}
func (c *Client) GetTradeReports() ([][]string, bool) {
	var result [][]string
	for {
		trades, err := c.ReadSpotUserTrade()
		if err != nil {
			break
		}
		for _, trade := range trades {
			var isMaker string
			if trade.IsMaker {
				isMaker = "true"
			} else {
				isMaker = "false"
			}
			st := decimal.NewFromInt(trade.TimeStamp.Unix()).String()
			data := []string{trade.Oid, trade.Symbol, "spot", c.subaccount, trade.Price.String(), trade.Qty.String(), trade.Side, strings.ToLower(trade.OrderType), trade.Fee.String(), trade.Qty.String(), st, isMaker}
			result = append(result, data)
		}
	}
	if len(result) == 0 {
		return result, false
	}
	return result, true
}
