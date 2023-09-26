package period

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/yzimhao/trading_engine"
	"github.com/yzimhao/trading_engine/utils"
	"github.com/yzimhao/trading_engine/utils/filecache"
	"xorm.io/xorm"
)

type Period struct {
	Symbol        string                     `xorm:"-" json:"-"`
	OpenAt        utils.Time                 `xorm:"notnull timestamp unique(open_at) default CURRENT_TIMESTAMP" json:"open_at"` //开盘时间
	CloseAt       utils.Time                 `xorm:"notnull timestamp default CURRENT_TIMESTAMP" json:"close_at"`                // 收盘时间
	Open          string                     `xorm:"decimal(30, 10) notnull" json:"open"`                                        //开盘价
	High          string                     `xorm:"decimal(30, 10) notnull" json:"high"`                                        // 最高价
	Low           string                     `xorm:"decimal(30, 10) notnull" json:"low"`                                         //最低价
	Close         string                     `xorm:"decimal(30, 10) notnull" json:"close"`                                       //收盘价(当前K线未结束的即为最新价)
	Volume        string                     `xorm:"decimal(30, 10) notnull" json:"volume"`                                      //成交量
	Amount        string                     `xorm:"decimal(30, 10) notnull" json:"amount"`                                      //成交额
	CreateTime    utils.Time                 `xorm:"timestamp created" json:"-"`
	UpdateTime    utils.Time                 `xorm:"timestamp updated" json:"-"`
	Interval      PeriodType                 `xorm:"-" json:"-"`
	raw           trading_engine.TradeResult `xorm:"-" json:"-"`
	lastOpenTime  int64                      `xorm:"-" json:"last_open_time"`
	lastCloseTime int64                      `xorm:"-" json:"last_close_time"`
}

func NewPeriod(symbol string, p PeriodType, tl trading_engine.TradeResult) *Period {
	tradetime := time.Unix(int64(tl.TradeTime/1e9), 0)
	logrus.Infof("trade_time: %s", utils.Time(tradetime).Format())
	open_at, close_at := get_start_end_time(tradetime, p)

	db := filecache.NewStorage(viper.GetString("haoquote.storage_path"), 10)
	defer db.Close()

	data := Period{}
	key := fmt.Sprintf("period_%d_%d", open_at.Unix(), close_at.Unix())
	cache, _ := db.Get("period", key)
	json.Unmarshal(cache, &data)

	//将计算结果写入缓存
	defer func() {
		logrus.Infof("data: %#v", data)
		raw, err := json.Marshal(data)
		logrus.Infof("set cache: %s %s", raw, err)
		db.Set("period", key, raw)
	}()

	data.raw = tl
	data.Interval = p
	data.Symbol = symbol
	data.OpenAt = utils.Time(open_at)
	data.CloseAt = utils.Time(close_at)

	data.get_open()
	data.get_high()
	data.get_low()
	data.get_close()
	data.get_volume()
	data.get_amount()

	return &data
}

func (p *Period) TableName() string {
	return fmt.Sprintf("period_%s_%s", p.Symbol, p.Interval)
}

func (p *Period) CreateTable(db *xorm.Engine) error {
	if p.Symbol == "" || p.Interval == "" {
		return fmt.Errorf("symbol or period is null")
	}

	exist, err := db.IsTableExist(p.TableName())
	if err != nil {
		return err
	}

	if !exist {
		err := db.CreateTables(p)
		if err != nil {
			return err
		}
		err = db.CreateIndexes(p)
		if err != nil {
			return err
		}
		err = db.CreateUniques(p)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Period) get_open() {
	if p.Open == "" {
		p.Open = p.raw.TradePrice.String()
		p.High = p.raw.TradePrice.String()
		p.Low = p.raw.TradePrice.String()
		p.Close = p.raw.TradePrice.String()
		p.Volume = p.raw.TradeQuantity.String()
		p.Amount = p.raw.TradePrice.Mul(p.raw.TradeQuantity).String()
	}
	if p.raw.TradeTime < p.lastOpenTime {
		p.Open = p.raw.TradePrice.String()
		p.lastOpenTime = p.raw.TradeTime
	}
}

func (p *Period) get_high() {
	if p.raw.TradePrice.Cmp(d(p.High)) > 0 {
		p.High = p.raw.TradePrice.String()
	}
}

func (p *Period) get_low() {
	if p.raw.TradePrice.Cmp(d(p.Low)) < 0 {
		p.Low = p.raw.TradePrice.String()
	}
}

func (p *Period) get_close() {
	if p.raw.TradeTime > p.lastCloseTime {
		p.Close = p.raw.TradePrice.String()
		p.lastCloseTime = p.raw.TradeTime
	}
}

func (p *Period) get_volume() {
	v := d(p.Volume).Add(p.raw.TradeQuantity)
	p.Volume = v.String()
}

func (p *Period) get_amount() {
	v := d(p.Amount).Add(p.raw.TradePrice.Mul(p.raw.TradeQuantity))
	p.Amount = v.String()
}

func d(a string) decimal.Decimal {
	b, _ := decimal.NewFromString(a)
	return b
}
