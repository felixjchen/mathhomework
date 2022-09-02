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
	wethPools := uniswap.FilterPools(wethFilter, allPools)
	fmt.Println(wethPools)
}
