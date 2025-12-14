package core

type TradeRecord struct {
	ID         int     `json:"id"`
	Timestamp  string  `json:"timestamp"`
	Asset      string  `json:"asset"`
	Operation  string  `json:"operation"`
	AmountTHB  float64 `json:"amount_thb"`
	CoinAmount float64 `json:"coin_amount"`
	Price      float64 `json:"price"`
	Deviation  float64 `json:"deviation"`
}

type AssetData struct {
	Asset        string  `json:"asset"`
	CurrentPrice float64 `json:"current_price"`
	CoinBalance  float64 `json:"coin_balance"`
	BalanceTHB   float64 `json:"balance_thb"`
	ActualPct    float64 `json:"actual_pct"`
	TargetPct    float64 `json:"target_pct"`
}

type PortfolioSummary struct {
	TotalValue float64
	ROI        float64
	Portfolio  []AssetData
}

var walletResp struct {
	Error  float64            `json:"error"`
	Result map[string]float64 `json:"result"`
}
