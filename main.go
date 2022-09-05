package main

import (
	"arbitrage_go/blockchain"
	"arbitrage_go/config"
	"arbitrage_go/uniswap"
	"fmt"
	"log"
	"math/big"

	"github.com/chenzhijie/go-web3/types"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

func wethFilter(i uniswap.Pool) bool {
	weth := config.Get().WETH_ADDRESS
	return i.Token0 == weth || i.Token1 == weth
}
func tokenBlacklistFilter(i uniswap.Pool) bool {
	_, token0Blacklisted := config.TOKEN_BLACKLIST[i.Token0]
	_, token1Blacklisted := config.TOKEN_BLACKLIST[i.Token1]
	return !token0Blacklisted && !token1Blacklisted
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	sugar := logger.Sugar()

	web3 := blockchain.GetWeb3()

	allPools := uniswap.GetAllPools()
	allPools = uniswap.FilterPools(tokenBlacklistFilter, allPools)
	sugar.Info("Got all pools")

	// pathing
	// adjacency list
	tokensToPools := make(map[common.Address][]uniswap.Pool)
	for _, pool := range allPools {
		tokensToPools[pool.Token0] = append(tokensToPools[pool.Token0], pool)
		tokensToPools[pool.Token1] = append(tokensToPools[pool.Token1], pool)
	}
	sugar.Info("Created Graph")

	// pool has weth
	wethPools := uniswap.FilterPools(wethFilter, allPools)

	// two hop
	weth := config.Get().WETH_ADDRESS
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
	sugar.Info("Found all 2-hops")

	poolToReserves := uniswap.UpdateReservesForPools(wethPools)
	sugar.Info("Updated Reserves")

	// poolToReserves := make(map[uniswap.Pool]uniswap.Reserve)

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
		arbProfit := big.NewInt(0).Sub(wethOut, wethIn)
		// sugar.Info(path, arbProfit)
		// sugar.Info(poolToReserves[path[0]], poolToReserves[path[1]])
		arbProfitMinusGas := big.NewInt(0).Sub(arbProfit, big.NewInt(5286645002416752))
		// sugar.Info(path, arbProfit)
		if arbProfitMinusGas.Sign() == 1 {
			sugar.Info("PROFIT", path, arbProfit)
			// profitablePathes = append(profitablePathes, path)

			// INEFFICIENT
			// pool, err := web3.Eth.NewContract(config.PAIR_ABI, path[0].Address.String())
			// if err != nil {
			// 	panic(err)
			// }
			// Build first txn
			amount0OutFirst := big.NewInt(0)
			amount1OutFirst := intermediateAmount
			if path[0].Token1 == weth {
				amount0OutFirst = intermediateAmount
				amount1OutFirst = big.NewInt(0)
			}
			firstTarget := common.Address(path[0].Address)
			// firstData, err := pool.EncodeABI("swap", amount0Out, amount1Out, path[1].Address, []byte{})
			// if err != nil {
			// 	panic(err)
			// }
			// sugar.Info(hex.EncodeToString(firstData))
			// sugar.Info(firstTarget)
			// sugar.Info(amount0Out, amount1Out, common.HexToAddress(config.BUNDLE_EXECUTOR_ADDRESS), []byte{})

			// pool2, err := web3.Eth.NewContract(config.PAIR_ABI, path[1].Address.String())
			// if err != nil {
			// 	panic(err)
			// }
			// build second txn
			amount0OutSecond := wethOut
			amount1OutSecond := big.NewInt(0)
			if path[1].Token1 == weth {
				amount0OutSecond = big.NewInt(0)
				amount1OutSecond = wethOut
			}
			secondTarget := common.Address(path[1].Address)
			// secondData, err := pool2.EncodeABI("swap", amount0Out, amount1Out, config.Get().BUNDLE_EXECUTOR_ADDRESS, []byte{})
			// if err != nil {
			// 	panic(err)
			// }

			// run bundle
			executor, err := web3.Eth.NewContract(config.BUNDLE_EXECTOR_ABI, config.Get().BUNDLE_EXECUTOR_ADDRESS.Hex())
			if err != nil {
				panic(err)
			}

			// sugar.Info(wethIn, big.NewInt(0), [1]common.Address{firstTarget}, [1][]byte{firstData})

			// data, err := executor.EncodeABI("uniswapWeth", wethIn, big.NewInt(0), [2]common.Address{firstTarget, secondTarget}, [2][]byte{firstData, secondData})
			// if err != nil {
			// 	panic(err)
			// }

			data, err := executor.EncodeABI("twohop", wethIn, [2]common.Address{firstTarget, secondTarget}, []*big.Int{amount0OutFirst, amount0OutSecond}, []*big.Int{amount1OutFirst, amount1OutSecond})
			if err != nil {
				panic(err)
			}
			// TODO Fail simulation
			// TODO Gas estimation
			call := &types.CallMsg{
				From: web3.Eth.Address(),
				To:   executor.Address(),
				Data: data,
				Gas:  types.NewCallMsgBigInt(big.NewInt(types.MAX_GAS_LIMIT)),
			}
			// fmt.Printf("call %v\n", call)
			gasLimit, err := web3.Eth.EstimateGas(call)
			if err != nil {
				// panic(err)
				sugar.Info(err)
			}
			fmt.Printf("Estimate gas limit %v\n", gasLimit)

			// TODO not sync
			// tx, err := web3.Eth.SyncSendEIP1559RawTransaction(
			// 	executor.Address(),
			// 	big.NewInt(0),
			// 	1010000,
			// 	web3.Utils.ToGWei(40),
			// 	web3.Utils.ToGWei(325),
			// 	data,
			// )
			// if err != nil {
			// 	panic(err)
			// }
			// if err == nil {
			// 	fmt.Printf("tx hash %v\n", tx.TxHash)
			// }
		}
	}
}
