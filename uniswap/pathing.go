package uniswap

import (
	"arbitrage_go/config"
	"arbitrage_go/util"

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

// func dfs(candidate Cycle, pairGraph map[Pair][]common.Address, maxPathLength int, visited map[common.Address]bool, cycles *[]Cycle) {
// 	head := candidate.Edges[len(candidate.Edges)-1]

// 	if _, exist := visited[head.Address]; !exist {
// 		visited[head.Address] = true

// 		isCycle := candidate.Tokens[0] == candidate.Tokens[len(candidate.Tokens)-1]
// 		if isCycle {
// 			*cycles = append(*cycles, candidate)
// 		} else {
// 			if len(candidate.Edges) < maxPathLength {
// 			}
// 		}
// 	}
// }

// // Pairs are nodes
// // Tokens are edges
// func GetCycles2(start common.Address, pairGraph map[Pair][]common.Address, graph map[common.Address][]Pair, maxPathLength int) []Cycle {
// 	cycles := []Cycle{}
// 	visited := make(map[common.Address]bool)

// 	// muCycles := sync.Mutex{}
// 	// muVisited := sync.Mutex{}
// 	// wg := sync.WaitGroup{}

// 	for _, pair := range graph[start] {
// 		_, second := SortTokens(start, pair)
// 		initialCandidate := Cycle{[]common.Address{start, second}, []Pair{pair}}
// 		dfs(initialCandidate, pairGraph, maxPathLength-1, visited, &cycles)
// 	}

// 	return cycles
// }

func GetCycles(start common.Address, graph map[common.Address][]Pair, maxPathLength int) []Cycle {
	cycles := []Cycle{}

	// dfs this needs rework
	stack := []Cycle{}
	stack = append(stack, Cycle{[]common.Address{start}, []Pair{}})
	for len(stack) > 0 {
		// pop
		candidate := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		head := candidate.Tokens[len(candidate.Tokens)-1]
		if start == head && len(candidate.Edges) > 0 {
			cycles = append(cycles, candidate)
		} else {
			// keep going
			if len(candidate.Edges) < maxPathLength {
				usedEdges := make(map[Pair]bool)
				for _, pair := range candidate.Edges {
					usedEdges[pair] = true
				}
				for _, pair := range graph[head] {
					// No reusing edges
					if _, exist := usedEdges[pair]; !exist {
						newCandidate := Cycle{candidate.Tokens[:], candidate.Edges[:]}
						_, newHead := SortTokens(head, pair)
						newCandidate.Tokens = append(newCandidate.Tokens, newHead)
						newCandidate.Edges = append(newCandidate.Edges, pair)
						stack = append(stack, newCandidate)
					}
				}
			}
		}
	}

	return cycles
}

func SortTokens(start common.Address, pair Pair) (common.Address, common.Address) {
	s := util.Ternary(pair.Token0 == start, pair.Token0, pair.Token1)
	t := util.Ternary(pair.Token0 != start, pair.Token0, pair.Token1)
	return s, t
}
