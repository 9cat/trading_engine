package trading_engine

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type OrderType int
type OrderSide int

const (
	OrderTypeLimit          OrderType = 0
	OrderTypeMarket         OrderType = 1
	OrderTypeMarketQuantity OrderType = 2
	OrderTypeMarketAmount   OrderType = 3

	OrderSideBuy  OrderSide = 0
	OrderSideSell OrderSide = 1
)

func (os OrderSide) String() string {
	if os == OrderSideSell {
		return "ask"
	}
	return "bid"
}

func (ot OrderType) String() string {
	switch ot {
	case OrderTypeMarket:
		return "market"
	case OrderTypeMarketAmount:
		return "market_amount"
	case OrderTypeMarketQuantity:
		return "market_qty"
	default:
		return "limit"
	}
}

func FormatDecimal2String(d decimal.Decimal, digit int) string {
	f, _ := d.Float64()
	format := "%." + fmt.Sprintf("%d", digit) + "f"
	return fmt.Sprintf(format, f)
}

func quickSort(nums []string, asc_desc string) []string {
	if len(nums) <= 1 {
		return nums
	}

	spilt := nums[0]
	left := []string{}
	right := []string{}
	mid := []string{}

	for _, v := range nums {
		vv, _ := decimal.NewFromString(v)
		sp, _ := decimal.NewFromString(spilt)
		if vv.Cmp(sp) == -1 {
			left = append(left, v)
		} else if vv.Cmp(sp) == 1 {
			right = append(right, v)
		} else {
			mid = append(mid, v)
		}
	}

	left = quickSort(left, asc_desc)
	right = quickSort(right, asc_desc)

	if asc_desc == "asc" {
		return append(append(left, mid...), right...)
	} else {
		return append(append(right, mid...), left...)
	}

	//return append(append(left, mid...), right...)
}

func sortMap2Slice(m map[string]string, ask_bid OrderSide) [][2]string {
	res := [][2]string{}
	keys := []string{}
	for k, _ := range m {
		keys = append(keys, k)
	}

	if ask_bid == OrderSideSell {
		keys = quickSort(keys, "asc")
	} else {
		keys = quickSort(keys, "desc")
	}

	for _, k := range keys {
		res = append(res, [2]string{k, m[k]})
	}
	return res
}
