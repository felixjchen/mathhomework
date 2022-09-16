package uniswap

import (
	"arbitrage_go/config"
	"arbitrage_go/util"
	"sync"

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

func dfs(candidate *Cycle, graph *map[common.Address][]Pair, maxPathLength int, cycles *[]Cycle, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()

	n := len(candidate.Tokens)
	head := candidate.Tokens[n-1]
	isCycle := candidate.Tokens[0] == candidate.Tokens[n-1] && 2 < n
	if isCycle {
		mu.Lock()
		*cycles = append(*cycles, *candidate)
		mu.Unlock()
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
					go dfs(&newCandidate, graph, maxPathLength, cycles, wg, mu)
				}
			}
		}
	}
}

// Pairs are nodes
func GetCycles2(start common.Address, graph map[common.Address][]Pair, maxPathLength int) []Cycle {

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	cycles := []Cycle{}

	initialCandidate := Cycle{[]common.Address{start}, []Pair{}}
	wg.Add(1)
	go dfs(&initialCandidate, &graph, maxPathLength, &cycles, &wg, &mu)
	wg.Wait()
	return cycles
}

func GetCycles(start common.Address, graph map[common.Address][]Pair, maxPathLength int, checkChan chan Cycle) {
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

func SortTokens(start common.Address, pair Pair) (common.Address, common.Address) {
	s := util.Ternary(pair.Token0 == start, pair.Token0, pair.Token1)
	t := util.Ternary(pair.Token0 != start, pair.Token0, pair.Token1)
	return s, t
}
