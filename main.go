package main

import (
	"arbitrage_go/config"
	"arbitrage_go/uniswap"
	"fmt"
	"math/big"

	"github.com/chenzhijie/go-web3"
	"github.com/ethereum/go-ethereum/common"
)

func wethFilter(i uniswap.Pool) bool {
	weth := common.HexToAddress(config.WETH_ADDRESS)
	return i.Token0 == weth || i.Token1 == weth
}
func main() {

	MUMBAI_URL := "https://polygon-mumbai.infura.io/v3/1de294ccc0da4f2ab105c9770ab3b962"
	web3, err := web3.NewWeb3(MUMBAI_URL)
	if err != nil {
		panic(err)
	}

	allPools := uniswap.GetAllPools()

	// pool has weth
	wethPools := uniswap.FilterPools(wethFilter, allPools)

	// two hop
	// pathing
	tokensToPools := make(map[common.Address][]uniswap.Pool)
	for _, pool := range allPools {
		tokensToPools[pool.Token0] = append(tokensToPools[pool.Token0], pool)
		tokensToPools[pool.Token1] = append(tokensToPools[pool.Token1], pool)
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

			samePair := (pool2.Token0 == weth || pool2.Token1 == weth) && (pool2.Token0 == intermediateToken || pool2.Token1 == intermediateToken)
			if pool1.Address != pool2.Address && samePair {
				pathes = append(pathes, [2]uniswap.Pool{pool1, pool2})
			}
		}
	}

	// pricing
	reserves := uniswap.UpdateReservesForPools(wethPools)
	poolToReserves := make(map[uniswap.Pool]uniswap.Reserve)
	for i, pool := range wethPools {
		poolToReserves[pool] = reserves[i]
	}
	// fmt.Println(poolToReserves)

	// profitablePathes := [][2]uniswap.Pool{}
	// Simulate path
	for _, path := range pathes {
		wethIn := web3.Utils.ToWei(0.001)

		// price first hop
		wethReserve := poolToReserves[path[0]].Reserve0
		intermediateReserve := poolToReserves[path[0]].Reserve1
		if path[0].Token1 == weth {
			wethReserve = poolToReserves[path[0]].Reserve1
			intermediateReserve = poolToReserves[path[0]].Reserve0
		}
		intermediateAmount := uniswap.GetAmountOut(wethIn, wethReserve, intermediateReserve)

		// price second hop
		wethReserve = poolToReserves[path[1]].Reserve0
		intermediateReserve = poolToReserves[path[1]].Reserve1
		if path[1].Token1 == weth {
			wethReserve = poolToReserves[path[1]].Reserve1
			intermediateReserve = poolToReserves[path[1]].Reserve0
		}

		wethOut := uniswap.GetAmountOut(intermediateAmount, intermediateReserve, wethReserve)

		// fmt.Println(wethOut)

		if big.NewInt(0).Sub(wethOut, wethIn).Sign() == 1 {
			fmt.Println("PROFIT", path, big.NewInt(0).Sub(wethOut, wethIn))
		}

	}
}
