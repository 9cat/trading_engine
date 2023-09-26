package trading_engine

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

type TradeResult struct {
	Symbol        string          `json:"symbol"`
	AskOrderId    string          `json:"ask"`
	BidOrderId    string          `json:"bid"`
	TradeQuantity decimal.Decimal `json:"trade_quantity"`
	TradePrice    decimal.Decimal `json:"trade_price"`
	TradeTime     int64           `json:"trade_time"`  //纳秒级
	MarketDone    string          `json:"market_done"` //市价订单标记，用于结算时候取消剩余未成交的部分
}

func (t *TradeResult) Json() []byte {
	raw, _ := json.Marshal(&t)
	return raw
}
