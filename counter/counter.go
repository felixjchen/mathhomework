package counter

import "sync"

type TSCounter struct {
	count uint64
	mu    *sync.Mutex
}

func NewTSCounter(i uint64) *TSCounter {
	return &TSCounter{
		count: i,
		mu:    &sync.Mutex{},
	}
}

func (n *TSCounter) Get() uint64 {
	return n.count
}

func (n *TSCounter) Inc() {
	n.count++
}

func (n *TSCounter) Dec() {
	n.count--
}

func (n *TSCounter) TSGet() uint64 {
	n.Lock()
	defer n.Unlock()
	return n.count
}

func (n *TSCounter) TSInc() {
	n.Lock()
	defer n.Unlock()
	n.count++
}

func (n *TSCounter) TSDec() {
	n.Lock()
	defer n.Unlock()
	n.count--
}

func (n *TSCounter) Lock() {
	n.mu.Lock()
}
func (n *TSCounter) Unlock() {
	n.mu.Unlock()
}
