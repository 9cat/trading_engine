package www

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/yzimhao/trading_engine/haoquote/ws"
	"github.com/yzimhao/trading_engine/types"
)

var (
	symbols_depth depth_data
)

type depth_data struct {
	data         map[string]map[string][][2]string
	price_digit  map[string]int64
	qty_digit    map[string]int64
	latest_price map[string]string
	sync.Mutex
}

func (d *depth_data) limit(side string, symbol string, size int) [][2]string {
	d.Lock()
	defer d.Unlock()

	max := len(d.data[symbol][side])
	if size <= 0 || size > max {
		size = max
	}
	return d.data[symbol][side][0:size]
}

func (d *depth_data) update(symbol string, data map[string][][2]string) {
	d.Lock()
	defer d.Unlock()
	d.data[symbol] = data
}

func (d *depth_data) set_digit(symbol string, price_digit, qty_digit int64) {
	d.Lock()
	defer d.Unlock()
	d.price_digit[symbol] = price_digit
	d.qty_digit[symbol] = qty_digit
}

func (d *depth_data) get_digit(symbol string) (price_digit, qty_digit int64) {
	d.Lock()
	defer d.Unlock()
	return d.price_digit[symbol], d.qty_digit[symbol]
}

func (d *depth_data) set_latest_price(symbol string, price string) {
	d.Lock()
	defer d.Unlock()
	d.latest_price[symbol] = price
}

func (d *depth_data) get_latest_price(symbol string) string {
	d.Lock()
	defer d.Unlock()

	return d.latest_price[symbol]
}

func sub_symbol_depth() {
	symbols := viper.GetStringMap("symbol")

	symbols_depth.data = make(map[string]map[string][][2]string)
	symbols_depth.price_digit = make(map[string]int64)
	symbols_depth.qty_digit = make(map[string]int64)
	symbols_depth.latest_price = make(map[string]string)

	for k, attr := range symbols {
		symbol := strings.ToLower(k)
		price_digit := attr.(map[string]any)["price_digit"].(int64)
		qty_digit := attr.(map[string]any)["qty_digit"].(int64)

		go sub_depth(symbol, price_digit, qty_digit)
		go sub_latest_price(symbol)
	}
}

func sub_depth(symbol string, price_digit, qty_digit int64) {
	channel := types.FormatBroadcastDepth.Format(symbol)
	ctx := context.Background()

	pubsub := rdc.PSubscribe(ctx, channel)
	defer pubsub.Close()

	symbols_depth.set_digit(symbol, price_digit, qty_digit)

	for {
		select {
		case msg := <-pubsub.Channel():
			var data map[string][][2]string

			err := json.Unmarshal([]byte(msg.Payload), &data)
			if err != nil {
				logrus.Infof("subscribe: %s data: %s err: %s", channel, msg.Payload, err)
			}

			symbols_depth.update(symbol, data)
			// websocket前端推送
			push_websocket_depth(symbol)
		}
	}

}

func sub_latest_price(symbol string) {
	channel := types.FormatBroadcastLatestPrice.Format(symbol)
	ctx := context.Background()

	pubsub := rdc.PSubscribe(ctx, channel)
	defer pubsub.Close()

	last := int64(0)

	for {
		select {
		case msg := <-pubsub.Channel():
			var data types.ChannelLatestPrice

			err := json.Unmarshal([]byte(msg.Payload), &data)
			if err != nil {
				logrus.Infof("subscribe: %s data: %s err: %s", channel, msg.Payload, err)
			}

			// logrus.Infof("last: %d now: %d %s", last, data.T, data.Price)
			if data.T >= last {
				symbols_depth.set_latest_price(symbol, data.Price)
				last = data.T

				// websocket前端推送
				to := types.MsgLatestPrice.Format(symbol)
				ws.M.Broadcast <- ws.MsgBody{
					To: to,
					Response: ws.Response{
						Type: to,
						Body: map[string]any{
							"latest_price": data.Price,
							"at":           data.T,
						},
					},
				}

				//计算24H涨跌幅
				market_24h(symbol, data.Price)
			}

		}
	}
}

func push_websocket_depth(symbol string) {
	to := types.MsgDepth.Format(symbol)
	ws.M.Broadcast <- ws.MsgBody{
		To: to,
		Response: ws.Response{
			Type: to,
			Body: gin.H{
				"asks": symbols_depth.limit("asks", symbol, 10),
				"bids": symbols_depth.limit("bids", symbol, 10),
			},
		},
	}
}
