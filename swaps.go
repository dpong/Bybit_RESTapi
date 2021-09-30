package bybitapi

import (
	"net/http"
	"strings"
)

type LastInfoForSymbolResponse struct {
	RetCode int    `json:"ret_code"`
	RetMsg  string `json:"ret_msg"`
	ExtCode string `json:"ext_code"`
	ExtInfo string `json:"ext_info"`
	Result  []struct {
		Symbol                 string  `json:"symbol"`
		BidPrice               string  `json:"bid_price"`
		AskPrice               string  `json:"ask_price"`
		LastPrice              string  `json:"last_price"`
		LastTickDirection      string  `json:"last_tick_direction"`
		PrevPrice24H           string  `json:"prev_price_24h"`
		Price24HPcnt           string  `json:"price_24h_pcnt"`
		HighPrice24H           string  `json:"high_price_24h"`
		LowPrice24H            string  `json:"low_price_24h"`
		PrevPrice1H            string  `json:"prev_price_1h"`
		Price1HPcnt            string  `json:"price_1h_pcnt"`
		MarkPrice              string  `json:"mark_price"`
		IndexPrice             string  `json:"index_price"`
		OpenInterest           float64 `json:"open_interest"`
		OpenValue              string  `json:"open_value"`
		TotalTurnover          string  `json:"total_turnover"`
		Turnover24H            string  `json:"turnover_24h"`
		TotalVolume            float64 `json:"total_volume"`
		Volume24H              float64 `json:"volume_24h"`
		FundingRate            string  `json:"funding_rate"`
		PredictedFundingRate   string  `json:"predicted_funding_rate"`
		NextFundingTime        string  `json:"next_funding_time"`
		CountdownHour          float64 `json:"countdown_hour"`
		DeliveryFeeRate        string  `json:"delivery_fee_rate"`
		PredictedDeliveryPrice string  `json:"predicted_delivery_price"`
		DeliveryTime           string  `json:"delivery_time"`
	} `json:"result"`
	TimeNow string `json:"time_now"`
}

func (p *Client) LastInfoForSymbol(symbol string) (swaps *LastInfoForSymbolResponse, err error) {
	params := make(map[string]string)
	if symbol != "" {
		params["symbol"] = strings.ToUpper(symbol)
	}
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	res, err := p.sendRequest("swap", http.MethodGet, "/v2/public/tickers", body, &params, false)
	if err != nil {
		return nil, err
	}
	// in Close()
	err = decode(res, &swaps)
	if err != nil {
		return nil, err
	}
	return swaps, nil
}

type SwapsInfoResponse struct {
	RetCode int    `json:"ret_code"`
	RetMsg  string `json:"ret_msg"`
	ExtCode string `json:"ext_code"`
	ExtInfo string `json:"ext_info"`
	Result  []struct {
		Name           string  `json:"name"`
		Alias          string  `json:"alias"`
		Status         string  `json:"status"`
		BaseCurrency   string  `json:"base_currency"`
		QuoteCurrency  string  `json:"quote_currency"`
		PriceScale     float64 `json:"price_scale"`
		TakerFee       string  `json:"taker_fee"`
		MakerFee       string  `json:"maker_fee"`
		LeverageFilter struct {
			MinLeverage  float64 `json:"min_leverage"`
			MaxLeverage  float64 `json:"max_leverage"`
			LeverageStep string  `json:"leverage_step"`
		} `json:"leverage_filter"`
		PriceFilter struct {
			MinPrice string `json:"min_price"`
			MaxPrice string `json:"max_price"`
			TickSize string `json:"tick_size"`
		} `json:"price_filter"`
		LotSizeFilter struct {
			MaxTradingQty float64 `json:"max_trading_qty"`
			MinTradingQty float64 `json:"min_trading_qty"`
			QtyStep       float64 `json:"qty_step"`
		} `json:"lot_size_filter"`
	} `json:"result"`
	TimeNow string `json:"time_now"`
}

func (p *Client) SwapsInfo() (swaps *SwapsInfoResponse, err error) {
	res, err := p.sendRequest("swap", http.MethodGet, "/v2/public/symbols", nil, nil, false)
	if err != nil {
		return nil, err
	}
	// in Close()
	err = decode(res, &swaps)
	if err != nil {
		return nil, err
	}
	return swaps, nil
}

type LastFundingRateResponse struct {
	RetCode int    `json:"ret_code"`
	RetMsg  string `json:"ret_msg"`
	ExtCode string `json:"ext_code"`
	ExtInfo string `json:"ext_info"`
	Result  struct {
		Symbol               string  `json:"symbol"`
		FundingRate          float64 `json:"funding_rate"`
		FundingRateTimestamp string  `json:"funding_rate_timestamp"`
	} `json:"result"`
	TimeNow string `json:"time_now"`
}

func (p *Client) LastFundingRate(symbol string) (result *LastFundingRateResponse, err error) {
	params := make(map[string]string)
	if symbol != "" {
		params["symbol"] = strings.ToUpper(symbol)
	}
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	res, err := p.sendRequest("swap", http.MethodGet, "/public/linear/funding/prev-funding-rate", body, &params, false)
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

type SetAutoAddMarginResponse struct {
	RetCode          int         `json:"ret_code"`
	RetMsg           string      `json:"ret_msg"`
	ExtCode          string      `json:"ext_code"`
	ExtInfo          string      `json:"ext_info"`
	Result           interface{} `json:"result"`
	TimeNow          string      `json:"time_now"`
	RateLimitStatus  int         `json:"rate_limit_status"`
	RateLimitResetMs int64       `json:"rate_limit_reset_ms"`
	RateLimit        int         `json:"rate_limit"`
}

func (p *Client) SetAutoAddMargin(symbol, side string, AutoAdd bool) (result *SetAutoAddMarginResponse, err error) {
	params := make(map[string]interface{})
	params["symbol"] = strings.ToUpper(symbol)
	params["side"] = side
	params["auto_add_margin"] = AutoAdd
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	res, err := p.sendRequest("swap", http.MethodPost, "/private/linear/position/set-auto-add-margin", body, nil, true)
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

type SwapPositionsResponse struct {
	RetCode int    `json:"ret_code"`
	RetMsg  string `json:"ret_msg"`
	ExtCode string `json:"ext_code"`
	ExtInfo string `json:"ext_info"`
	Result  []struct {
		UserID              int    `json:"user_id"`
		Symbol              string `json:"symbol"`
		Side                string `json:"side"`
		Size                int    `json:"size"`
		PositionValue       int    `json:"position_value"`
		EntryPrice          int    `json:"entry_price"`
		LiqPrice            int    `json:"liq_price"`
		BustPrice           int    `json:"bust_price"`
		Leverage            int    `json:"leverage"`
		IsIsolated          bool   `json:"is_isolated"`
		AutoAddMargin       int    `json:"auto_add_margin"`
		PositionMargin      int    `json:"position_margin"`
		OccClosingFee       int    `json:"occ_closing_fee"`
		RealisedPnl         int    `json:"realised_pnl"`
		CumRealisedPnl      int    `json:"cum_realised_pnl"`
		FreeQty             int    `json:"free_qty"`
		TpSlMode            string `json:"tp_sl_mode"`
		UnrealisedPnl       int    `json:"unrealised_pnl"`
		DeleverageIndicator int    `json:"deleverage_indicator"`
		RiskID              int    `json:"risk_id"`
		StopLoss            int    `json:"stop_loss"`
		TakeProfit          int    `json:"take_profit"`
		TrailingStop        int    `json:"trailing_stop"`
	} `json:"result"`
	TimeNow          string `json:"time_now"`
	RateLimitStatus  int    `json:"rate_limit_status"`
	RateLimitResetMs int64  `json:"rate_limit_reset_ms"`
	RateLimit        int    `json:"rate_limit"`
}

func (p *Client) SwapPositions() (result *SwapPositionsResponse, err error) {
	res, err := p.sendRequest("swap", http.MethodGet, "/private/linear/position/list", nil, nil, true)
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

type SetLeverageResponse struct {
	RetCode          int         `json:"ret_code"`
	RetMsg           string      `json:"ret_msg"`
	ExtCode          string      `json:"ext_code"`
	ExtInfo          string      `json:"ext_info"`
	Result           interface{} `json:"result"`
	TimeNow          string      `json:"time_now"`
	RateLimitStatus  int         `json:"rate_limit_status"`
	RateLimitResetMs int64       `json:"rate_limit_reset_ms"`
	RateLimit        int         `json:"rate_limit"`
}

func (p *Client) SetLeverage(symbol string, leverage int) (result *SetLeverageResponse, err error) {
	params := make(map[string]interface{})
	params["symbol"] = strings.ToUpper(symbol)
	params["buy_leverage"] = leverage
	params["sell_leverage"] = leverage
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	res, err := p.sendRequest("swap", http.MethodPost, "/private/linear/position/set-leverage", body, nil, true)
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

type GetRiskLimitResponse struct {
	RetCode int    `json:"ret_code"`
	RetMsg  string `json:"ret_msg"`
	ExtCode string `json:"ext_code"`
	ExtInfo string `json:"ext_info"`
	Result  []struct {
		ID             int      `json:"id"`
		Symbol         string   `json:"symbol"`
		Limit          float64  `json:"limit"`
		MaintainMargin float64  `json:"maintain_margin"`
		StartingMargin float64  `json:"starting_margin"`
		Section        []string `json:"section"`
		IsLowestRisk   float64  `json:"is_lowest_risk"`
		CreatedAt      string   `json:"created_at"`
		UpdatedAt      string   `json:"updated_at"`
		MaxLeverage    float64  `json:"max_leverage"`
	} `json:"result"`
	TimeNow string `json:"time_now"`
}

func (p *Client) GetRiskLimit(symbol string) (result *GetRiskLimitResponse, err error) {
	params := make(map[string]string)
	params["symbol"] = strings.ToUpper(symbol)
	res, err := p.sendRequest("swap", http.MethodGet, "/public/linear/risk-limit", nil, &params, false)
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

type GetWalletBalanceResponse struct {
	RetCode int    `json:"ret_code"`
	RetMsg  string `json:"ret_msg"`
	ExtCode string `json:"ext_code"`
	ExtInfo string `json:"ext_info"`
	Result  struct {
		Btc struct {
			Equity           float64 `json:"equity"`
			AvailableBalance float64 `json:"available_balance"`
			UsedMargin       float64 `json:"used_margin"`
			OrderMargin      float64 `json:"order_margin"`
			PositionMargin   float64 `json:"position_margin"`
			OccClosingFee    float64 `json:"occ_closing_fee"`
			OccFundingFee    float64 `json:"occ_funding_fee"`
			WalletBalance    float64 `json:"wallet_balance"`
			RealisedPnl      float64 `json:"realised_pnl"`
			UnrealisedPnl    float64 `json:"unrealised_pnl"`
			CumRealisedPnl   float64 `json:"cum_realised_pnl"`
			GivenCash        float64 `json:"given_cash"`
			ServiceCash      float64 `json:"service_cash"`
		} `json:"BTC"`
	} `json:"result"`
	TimeNow          string `json:"time_now"`
	RateLimitStatus  int    `json:"rate_limit_status"`
	RateLimitResetMs int64  `json:"rate_limit_reset_ms"`
	RateLimit        int    `json:"rate_limit"`
}

func (p *Client) GetWalletBalance() (result *GetWalletBalanceResponse, err error) {
	res, err := p.sendRequest("swap", http.MethodGet, "/v2/private/wallet/balance", nil, nil, true)
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
