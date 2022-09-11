package uniswap

import (
	"arbitrage_go/config"

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
