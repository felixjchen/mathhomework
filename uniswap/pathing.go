package uniswap

import (
	"arbitrage_go/util"

	"github.com/ethereum/go-ethereum/common"
)

// func dfs(candidate *Cycle, graph *map[common.Address][]Pair, maxPathLength int, wg *sync.WaitGroup, checkChan chan Cycle) {
// 	defer wg.Done()

// 	n := len(candidate.Tokens)
// 	head := candidate.Tokens[n-1]
// 	isCycle := candidate.Tokens[0] == candidate.Tokens[n-1] && 2 < n
// 	if isCycle {
// 		checkChan <- *candidate
// 	} else {
// 		if n-1 < maxPathLength {
// 			for _, pair := range (*graph)[head] {
// 				// No reusing edges
// 				if !util.Contains(candidate.Edges, pair) {
// 					_, newHead := SortTokens(head, pair)
// 					newTokens := append([]common.Address{}, candidate.Tokens...)
// 					newEdges := append([]Pair{}, candidate.Edges...)
// 					newTokens = append(newTokens, newHead)
// 					newEdges = append(newEdges, pair)
// 					newCandidate := Cycle{newTokens, newEdges}
// 					wg.Add(1)
// 					go dfs(&newCandidate, graph, maxPathLength, wg, checkChan)
// 				}
// 			}
// 		}
// 	}
// }

// // Pairs are nodes
// func GetCycles2(start common.Address, graph map[common.Address][]Pair, maxPathLength int, checkChan chan Cycle) {

// 	wg := sync.WaitGroup{}
// 	// mu := sync.Mutex{}

// 	initialCandidate := Cycle{[]common.Address{start}, []Pair{}}
// 	wg.Add(1)
// 	go dfs(&initialCandidate, &graph, maxPathLength, &wg, checkChan)
// 	wg.Wait()
// }

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

func SortTokens(start common.Address, pair Pair) (common.Address, common.Address) {
	s := util.Ternary(pair.Token0 == start, pair.Token0, pair.Token1)
	t := util.Ternary(pair.Token0 != start, pair.Token0, pair.Token1)
	return s, t
}
