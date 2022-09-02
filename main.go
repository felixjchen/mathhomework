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
	web3.Eth.SetChainId(80001)

	err = web3.Eth.SetAccount("ea0d86ce7b7c394ca92cafadb8c8b50e82820d79de32f993a78b16c0ab5b73ad")
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

		if big.NewInt(0).Sub(wethOut, wethIn).Sign() == 1 {
			fmt.Println("PROFIT", path, big.NewInt(0).Sub(wethOut, wethIn))
			// profitablePathes = append(profitablePathes, path)

			// INEFFICIENT
			pool, err := web3.Eth.NewContract(config.PAIR_ABI, path[0].Address.String())
			if err != nil {
				panic(err)
			}
			// Build first txn
			amount0Out := big.NewInt(0)
			amount1Out := intermediateAmount
			if path[0].Token1 == weth {
				amount0Out = intermediateAmount
				amount1Out = big.NewInt(0)
			}
			firstTarget := common.Address(path[0].Address)
			firstData, err := pool.EncodeABI("swap", amount0Out, amount1Out, path[1].Address, []byte{})
			if err != nil {
				panic(err)
			}
			// fmt.Println(hex.EncodeToString(firstData))
			// fmt.Println(firstTarget)
			// fmt.Println(amount0Out, amount1Out, common.HexToAddress(config.BUNDLE_EXECUTOR_ADDRESS), []byte{})

			pool2, err := web3.Eth.NewContract(config.PAIR_ABI, path[1].Address.String())
			if err != nil {
				panic(err)
			}
			// build second txn
			amount0Out = wethOut
			amount1Out = big.NewInt(0)
			if path[1].Token1 == weth {
				amount0Out = big.NewInt(0)
				amount1Out = wethOut
			}
			secondTarget := common.Address(path[1].Address)
			secondData, err := pool2.EncodeABI("swap", amount0Out, amount1Out, common.HexToAddress(config.BUNDLE_EXECUTOR_ADDRESS), []byte{})
			if err != nil {
				panic(err)
			}

			// run bundle
			executor, err := web3.Eth.NewContract(config.BUNDLE_EXECTOR_ABI, config.BUNDLE_EXECUTOR_ADDRESS)
			if err != nil {
				panic(err)
			}

			// fmt.Println(wethIn, big.NewInt(0), [1]common.Address{firstTarget}, [1][]byte{firstData})

			data, err := executor.EncodeABI("uniswapWeth", wethIn, big.NewInt(0), [2]common.Address{firstTarget, secondTarget}, [2][]byte{firstData, secondData})
			if err != nil {
				panic(err)
			}

			tx, err := web3.Eth.SyncSendEIP1559RawTransaction(
				executor.Address(),
				big.NewInt(0),
				1010000,
				web3.Utils.ToGWei(25),
				web3.Utils.ToGWei(325),
				data,
			)
			if err != nil {
				panic(err)
			}
			fmt.Printf("tx hash %v\n", tx.TxHash)
		}
	}
}
