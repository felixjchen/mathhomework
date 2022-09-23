package uniswap

import (
	"arbitrage_go/config"
	"arbitrage_go/logging"
	"arbitrage_go/util"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// type Cycle struct {
// 	pools []Pool
// }

func GetTwoHops(tokensToPools map[common.Address][]Pair) [][2]Pair {
	weth := config.Get().WETH_ADDRESS
	pathes := [][2]Pair{}
	for _, pool1 := range tokensToPools[weth] {
		intermediateToken := pool1.Token0
		if pool1.Token0 == weth {
			intermediateToken = pool1.Token1
		}
		for _, pool2 := range tokensToPools[intermediateToken] {
			samePair := (pool2.Token0 == weth || pool2.Token1 == weth) && (pool2.Token0 == intermediateToken || pool2.Token1 == intermediateToken)
			if pool1.Address != pool2.Address && samePair {
				pathes = append(pathes, [2]Pair{pool1, pool2})
			}
		}
	}
	return pathes
}

type Cycle struct {
	Tokens []common.Address
	Edges  []Pair
}

func dfs(candidate *Cycle, graph *map[common.Address][]Pair, maxPathLength int, wg *sync.WaitGroup, checkChan chan Cycle) {
	defer wg.Done()

	n := len(candidate.Tokens)
	head := candidate.Tokens[n-1]
	isCycle := candidate.Tokens[0] == candidate.Tokens[n-1] && 2 < n
	if isCycle {
		checkChan <- *candidate
	} else {
		if n-1 < maxPathLength {
			for _, pair := range (*graph)[head] {
				// No reusing edges
				if !util.Contains(candidate.Edges, pair) {
					_, newHead := SortTokens(head, pair)
					newTokens := append([]common.Address{}, candidate.Tokens...)
					newEdges := append([]Pair{}, candidate.Edges...)
					newTokens = append(newTokens, newHead)
					newEdges = append(newEdges, pair)
					newCandidate := Cycle{newTokens, newEdges}
					wg.Add(1)
					go dfs(&newCandidate, graph, maxPathLength, wg, checkChan)
				}
			}
		}
	}
}

// Pairs are nodes
func GetCycles2(start common.Address, graph map[common.Address][]Pair, maxPathLength int, checkChan chan Cycle) {

	wg := sync.WaitGroup{}
	// mu := sync.Mutex{}

	initialCandidate := Cycle{[]common.Address{start}, []Pair{}}
	wg.Add(1)
	go dfs(&initialCandidate, &graph, maxPathLength, &wg, checkChan)
	wg.Wait()
}

func GetCyclesToChan(start common.Address, graph map[common.Address][]Pair, maxPathLength int, checkChan chan Cycle) {
	// dfs
	stack := []Cycle{}
	stack = append(stack, Cycle{[]common.Address{start}, []Pair{}})
	for len(stack) > 0 {
		// pop
		candidate := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		head := candidate.Tokens[len(candidate.Tokens)-1]
		if start == head && len(candidate.Edges) > 0 {
			checkChan <- candidate
		} else {
			// keep going
			if len(candidate.Edges) < maxPathLength {
				for _, pair := range graph[head] {
					// No reusing edges
					if !util.Contains(candidate.Edges, pair) {
						_, newHead := SortTokens(head, pair)
						newTokens := append(append([]common.Address{}, candidate.Tokens...), newHead)
						newEdges := append(append([]Pair{}, candidate.Edges...), pair)
						newCandidate := Cycle{newTokens, newEdges}
						stack = append(stack, newCandidate)
					}
				}
			}
		}
	}
}

func GetCyclesArray(start common.Address, graph map[common.Address][]Pair, maxPathLength int) []Cycle {
	res := []Cycle{}
	mu := sync.Mutex{}
	sugar := logging.GetSugar()
	go func() {
		for {
			time.Sleep(time.Second)
			mu.Lock()
			sugar.Info("Finding cycles, found: ", len(res))
			mu.Unlock()
		}
	}()

	// dfs
	stack := []Cycle{}
	stack = append(stack, Cycle{[]common.Address{start}, []Pair{}})
	for len(stack) > 0 {
		// pop
		candidate := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		head := candidate.Tokens[len(candidate.Tokens)-1]
		if start == head && len(candidate.Edges) > 0 {
			mu.Lock()
			res = append(res, candidate)
			if len(res) > 1082793 {
				return res
			}
			mu.Unlock()
		} else {
			// keep going
			if len(candidate.Edges) < maxPathLength {
				for _, pair := range graph[head] {
					// No reusing edges
					if !util.Contains(candidate.Edges, pair) {
						_, newHead := SortTokens(head, pair)
						newTokens := append(append([]common.Address{}, candidate.Tokens...), newHead)
						newEdges := append(append([]Pair{}, candidate.Edges...), pair)
						newCandidate := Cycle{newTokens, newEdges}
						stack = append(stack, newCandidate)
					}
				}
			}
		}
	}
	return res
}

func SortTokens(start common.Address, pair Pair) (common.Address, common.Address) {
	s := util.Ternary(pair.Token0 == start, pair.Token0, pair.Token1)
	t := util.Ternary(pair.Token0 != start, pair.Token0, pair.Token1)
	return s, t
}
