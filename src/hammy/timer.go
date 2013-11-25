package hammy

import (
	"container/heap"
	"time"
)

type hostTimer struct {
	Host      string
	NextCheck time.Time
}

type hostHeap []hostTimer

func (h hostHeap) Len() int           { return len(h) }
func (h hostHeap) Less(i, j int) bool { return h[i].NextCheck.Before(h[j].NextCheck) }
func (h hostHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *hostHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(hostTimer))
}

func (h *hostHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type Timer struct {
	Proc Processor

	hHeap hostHeap
}

func NewTimer() *Timer {
	t := &Timer{}
	heap.Init(t.hHeap)
	return t
}

func (t *Timer) UpdateList() {

}

func (t *Timer) Run() {
	for {
		// Достаем голову кучи

		// Ставим в очередь

		// Повторяем пока есть готовые к выполнению задачи

		time.Sleep(time.Second)
	}
}
