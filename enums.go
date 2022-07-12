package bybitapi

const (
	Buy      = "Buy"
	Sell     = "Sell"
	GTC      = "GoodTillCancel"
	FOK      = "FillOrKill"
	IOC      = "ImmediateOrCancel"
	PostOnly = "PostOnly"
	Limit    = "Limit"
	Market   = "Market"

	ProductPerp = "perp"
	ProductSpot = "spot"
	// for spot
	SpotGTC        = "GTC"
	SpotFOK        = "FOK"
	SpotIOC        = "IOC"
	SpotLimit      = "LIMIT"
	SpotMARKET     = "MARKET"
	SpotLimitMaker = "LIMIT_MAKER"

	// userTrade
	UserTradeBuy         = "buy"
	UserTradeSell        = "sell"
	Filled               = "FILLED"
	PartialFilled        = "PARTIALLY_FILLED"
	OrderTypeLimit       = "LIMIT"
	OrderTypeMarketQuote = "MARKET_OF_QUOTE"
	OrderTypeMarketBase  = "MARKET_OF_BASE"
)
