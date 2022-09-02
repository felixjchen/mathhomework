package main

import (
	"arbitrage_go/config"
	"arbitrage_go/uniswap"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

func wethFilter(i uniswap.Pool) bool {
	weth := common.HexToAddress(config.WETH_ADDRESS)
	return i.Token0 == weth || i.Token1 == weth
}
func main() {
	allPools := uniswap.GetAllPools()

	// pool has weth
	wethPools := uniswap.FilterPools(wethFilter, allPools)

	// two hop
	// pathing
	tokensToPools := make(map[common.Address][]uniswap.Pool)
	for _, pool := range wethPools {
		tokensToPools[pool.Token0] = append(tokensToPools[pool.Token0], pool)
		tokensToPools[pool.Token1] = append(tokensToPools[pool.Token0], pool)
	}

	// { WETH : [A , B , C]}
	// GLOBAL WETH
	weth := common.HexToAddress(config.WETH_ADDRESS)
	pathes := [][2]uniswap.Pool{}
	for _, pool1 := range tokensToPools[weth] {
		intermediateToken := pool1.Token0
		if pool1.Token0 == weth {
			intermediateToken = pool1.Token1
		}
		for _, pool2 := range tokensToPools[intermediateToken] {
			if pool1.Address != pool2.Address {
				pathes = append(pathes, [2]uniswap.Pool{pool1, pool2})
			}
		}
	}

	fmt.Println(pathes)
	// pricing
	// reserves := uniswap.UpdateReservesForPools(wethPools)
	// fmt.Println(reserves)
}
