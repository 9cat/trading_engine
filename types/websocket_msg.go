package types

import "strings"

type WebSocketMsgType string
type WebSocketKlineMsgType string

const (
	MsgDepth       WebSocketMsgType = "depth.{symbol}"
	MsgTrade       WebSocketMsgType = "tradelog.{symbol}"
	MsgLatestPrice WebSocketMsgType = "latest_price.{symbol}"

	MsgMarketKLine WebSocketKlineMsgType = "kline.{period}.{symbol}"
	// MsgMarketKLineM3  WebSocketMsgType = "kline.m3.{symbol}"
	// MsgMarketKLineM5  WebSocketMsgType = "kline.m5.{symbol}"
	// MsgMarketKLineM15 WebSocketMsgType = "kline.m15.{symbol}"
	// MsgMarketKLineM30 WebSocketMsgType = "kline.m30.{symbol}"
	// MsgMarketKLineH1  WebSocketMsgType = "kline.h1.{symbol}"
	// MsgMarketKLineH2  WebSocketMsgType = "kline.h2.{symbol}"
	// MsgMarketKLineH4  WebSocketMsgType = "kline.h4.{symbol}"
	// MsgMarketKLineH6  WebSocketMsgType = "kline.h6.{symbol}"
	// MsgMarketKLineH8  WebSocketMsgType = "kline.h8.{symbol}"
	// MsgMarketKLineH12 WebSocketMsgType = "kline.h12.{symbol}"
	// MsgMarketKLineD1  WebSocketMsgType = "kline.d1.{symbol}"
	// MsgMarketKLineD3  WebSocketMsgType = "kline.d3.{symbol}"
	// MsgMarketKLineW1  WebSocketMsgType = "kline.w1.{symbol}"
	// MsgMarketKLineMN  WebSocketMsgType = "kline.mn.{symbol}"
)

func (w WebSocketMsgType) Format(symbol string) string {
	key := strings.Replace(string(w), "{symbol}", symbol, -1)
	return key
}

func (w WebSocketKlineMsgType) Format(period, symbol string) string {
	key := strings.Replace(string(w), "{symbol}", symbol, -1)
	key = strings.Replace(key, "{period}", period, -1)
	return key
}
