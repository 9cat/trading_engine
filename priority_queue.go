package trading_engine

import (
	"container/heap"
	"sync"

	"github.com/shopspring/decimal"
)

type QueueItem interface {
	SetIndex(index int)
	SetQuantity(quantity decimal.Decimal)
	SetAmount(amount decimal.Decimal)
	Less(item QueueItem) bool
	GetIndex() int
	GetUniqueId() string
	GetPrice() decimal.Decimal
	GetQuantity() decimal.Decimal
	GetCreateTime() int64
	GetOrderSide() OrderSide
	GetOrderType() OrderType
	GetAmount() decimal.Decimal //订单金额，在市价订单的时候生效，限价单不需要这个字段
}

type PriorityQueue []QueueItem

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Less(pq[j])
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].SetIndex(i)
	pq[j].SetIndex(j)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.SetIndex(-1)
	*pq = old[0 : n-1]
	return item
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	x.(QueueItem).SetIndex(n)
	*pq = append(*pq, x.(QueueItem))
}

func NewQueue() *OrderQueue {
	pq := make(PriorityQueue, 0)
	heap.Init(&pq)

	queue := OrderQueue{
		pq: &pq,
		m:  make(map[string]*QueueItem),
	}
	return &queue
}

type OrderQueue struct {
	pq *PriorityQueue
	m  map[string]*QueueItem
	sync.Mutex

	depth [][2]string

	onEventUpdate func(QueueItem)
	onEventRemove func(QueueItem)
}

func (o *OrderQueue) Len() int {
	return o.pq.Len()
}

func (o *OrderQueue) Push(item QueueItem) (exist bool) {
	o.Lock()
	defer o.Unlock()

	id := item.GetUniqueId()
	if _, ok := o.m[id]; ok {
		return true
	}

	heap.Push(o.pq, item)
	o.m[id] = &item

	if o.onEventUpdate != nil {
		o.onEventUpdate(item)
	}
	return false
}

func (o *OrderQueue) Get(index int) QueueItem {
	n := o.pq.Len()
	if n <= index {
		return nil
	}

	return (*o.pq)[index]
}

func (o *OrderQueue) Top() QueueItem {
	return o.Get(0)
}

func (o *OrderQueue) Remove(uniqId string) QueueItem {
	o.Lock()
	defer o.Unlock()

	old, ok := o.m[uniqId]
	if !ok {
		return nil
	}

	item := heap.Remove(o.pq, (*old).GetIndex())
	delete(o.m, uniqId)

	if o.onEventRemove != nil {
		o.onEventRemove(item.(QueueItem))
	}
	return item.(QueueItem)
}

func (o *OrderQueue) SetQuantity(obj QueueItem, qty decimal.Decimal) QueueItem {
	obj.SetQuantity(qty)

	if o.onEventUpdate != nil {
		o.onEventUpdate(obj)
	}
	return obj
}

func (o *OrderQueue) clean() {
	o.Lock()
	defer o.Unlock()

	pq := make(PriorityQueue, 0)
	heap.Init(&pq)
	o.pq = &pq
	o.m = make(map[string]*QueueItem)
}
