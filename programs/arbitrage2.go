package programs

import (
	"arbitrage_go/blockchain"
	"arbitrage_go/config"
	"arbitrage_go/uniswap"
	"log"
	"math/big"

	"github.com/chenzhijie/go-web3/types"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

func Arbitrage2Main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	sugar := logger.Sugar()

	allPairs := uniswap.GetAllPairsArray()
	sugar.Info("Got ", len(allPairs), " pairs")

	// adjacency list
	// token -> [edges]
	graph := make(map[common.Address][]uniswap.Pair)
	for _, pair := range allPairs {
		graph[pair.Token0] = append(graph[pair.Token0], pair)
		graph[pair.Token1] = append(graph[pair.Token1], pair)
	}
	sugar.Info("Created Graph")

	// adjacency list
	// pair -> [token0, token1]
	pairGraph := make(map[uniswap.Pair][]common.Address)
	for _, pair := range allPairs {
		pairGraph[pair] = append(pairGraph[pair], pair.Token0)
		pairGraph[pair] = append(pairGraph[pair], pair.Token1)
	}
	sugar.Info("Created Pair Graph")

	cycles := uniswap.GetCycles(config.Get().WETH_ADDRESS, graph, 3)
	sugar.Info("Found ", len(cycles), " cycles")

	// Simulate path
	web3 := blockchain.GetWeb3()
	// weth := config.Get().WETH_ADDRESS
	pairToReserves := uniswap.GetReservesForPairs(allPairs)
	sugar.Info("Updated ", len(allPairs), " Reserves")
	for _, cycle := range cycles {
		E0, E1 := uniswap.GetE0E1ForCycle(cycle, pairToReserves)

		if new(big.Int).Sub(E0, E1).Sign() == -1 {
			amountIn := uniswap.GetOptimalAmountIn(E0, E1)

			if amountIn.Sign() == 1 {
				amountsOut := uniswap.GetAmountsOutCycle(pairToReserves, amountIn, cycle)
				arbProfit := big.NewInt(0).Sub(amountsOut[len(amountsOut)-1], amountIn)

				if arbProfit.Sign() == 1 {
					targets := uniswap.GetCycleTargets(cycle)
					amounts0Out, amounts1Out := uniswap.GetCycleAmountsOut(cycle, amountsOut)

					// run bundle
					executor, err := web3.Eth.NewContract(config.BUNDLE_EXECTOR_ABI, config.Get().BUNDLE_EXECUTOR_ADDRESS.Hex())
					if err != nil {
						panic(err)
					}
					// fmt.Println(cycle, amountsOut)
					// fmt.Println(amountIn, targets, amounts0Out, amounts1Out)

					data, err := executor.EncodeABI("hoppity", amountIn, targets, amounts0Out, amounts1Out)
					if err != nil {
						panic(err)
					}
					// TODO Gas estimation
					call := &types.CallMsg{
						From: web3.Eth.Address(),
						To:   executor.Address(),
						Data: data,
						Gas:  types.NewCallMsgBigInt(big.NewInt(types.MAX_GAS_LIMIT)),
					}
					gasLimit, err := web3.Eth.EstimateGas(call)
					if err != nil {
						sugar.Error(err)
					} else {
						sugar.Info("Estimate gas limit %v\n", gasLimit)
						// 0.006 ether
						gasGweiPrediction := big.NewInt(0)
						netProfit := new(big.Int).Sub(arbProfit, gasGweiPrediction)

						sugar.Info("Estimated Profit ", cycle.Edges, arbProfit, " SUB GAS ", netProfit)
						if netProfit.Sign() == 1 {
							// TODO not sync
							gasTipCap := web3.Utils.ToGWei(40)
							gasFeeCap := web3.Utils.ToGWei(325)
							tx, err := web3.Eth.SyncSendEIP1559RawTransaction(
								executor.Address(),
								new(big.Int),
								gasLimit+2000,
								gasTipCap,
								gasFeeCap,
								data,
							)
							if err != nil {
								panic(err)
							}
							if err == nil {
								sugar.Info("tx hash %v\n", tx.TxHash)
							}
						}
					}
				}
			}
		}
	}
}

// func Arbitrage2(path [2]uniswap.Pair, pairToReserves map[uniswap.Pair]uniswap.Reserve, gasTipCap *big.Int, gasFeeCap *big.Int) {
// 	// TODO_LOW clean up
// 	logger, err := zap.NewProduction()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	sugar := logger.Sugar()
// 	weth := config.Get().WETH_ADDRESS
// 	web3 := blockchain.GetWeb3()

// 	intermediateToken := util.Ternary(path[0].Token0 != weth, path[0].Token0, path[0].Token1)

// 	R0 := util.Ternary(path[0].Token0 == weth, pairToReserves[path[0]].Reserve0, pairToReserves[path[0]].Reserve1)
// 	R1 := util.Ternary(path[0].Token0 == intermediateToken, pairToReserves[path[0]].Reserve0, pairToReserves[path[0]].Reserve1)

// 	R1_ := util.Ternary(path[1].Token0 == intermediateToken, pairToReserves[path[1]].Reserve0, pairToReserves[path[1]].Reserve1)
// 	R2 := util.Ternary(path[1].Token0 == weth, pairToReserves[path[1]].Reserve0, pairToReserves[path[1]].Reserve1)

// 	E0, E1 := uniswap.GetE0E1(R0, R1, R1_, R2)
// 	fmt.Println(E0, E1)

// 	// if new(big.Int).Sub(E0, E1).Sign() == -1 {

// 	// 	wethIn := uniswap.GetOptimalWethIn(E0, E1)

// 	// 	// Min on current balance
// 	// 	// if big.NewInt(0).Sub(wethIn, big.NewInt(820000000000000000)).Sign() == 1 {
// 	// 	// 	wethIn = big.NewInt(820000000000000000)
// 	// 	// }

// 	// 	if wethIn.Sign() == 1 {
// 	// 		// price first hop
// 	// 		wethReserve := pairToReserves[path[0]].Reserve0
// 	// 		intermediateReserve := pairToReserves[path[0]].Reserve1
// 	// 		if path[0].Token1 == weth {
// 	// 			wethReserve = pairToReserves[path[0]].Reserve1
// 	// 			intermediateReserve = pairToReserves[path[0]].Reserve0
// 	// 		}
// 	// 		intermediateAmount := uniswap.GetAmountOut(wethIn, wethReserve, intermediateReserve)

// 	// 		// price second hop
// 	// 		wethReserve = pairToReserves[path[1]].Reserve0
// 	// 		intermediateReserve = pairToReserves[path[1]].Reserve1
// 	// 		if path[1].Token1 == weth {
// 	// 			wethReserve = pairToReserves[path[1]].Reserve1
// 	// 			intermediateReserve = pairToReserves[path[1]].Reserve0
// 	// 		}
// 	// 		wethOut := uniswap.GetAmountOut(intermediateAmount, intermediateReserve, wethReserve)

// 	// 		arbProfit := big.NewInt(0).Sub(wethOut, wethIn)

// 	// 		if arbProfit.Sign() == 1 {
// 	// 			// Build first txn
// 	// 			amount0OutFirst := new(big.Int)
// 	// 			amount1OutFirst := intermediateAmount
// 	// 			if path[0].Token1 == weth {
// 	// 				amount0OutFirst = intermediateAmount
// 	// 				amount1OutFirst = new(big.Int)
// 	// 			}
// 	// 			firstTarget := common.Address(path[0].Address)

// 	// 			// build second txn
// 	// 			amount0OutSecond := wethOut
// 	// 			amount1OutSecond := new(big.Int)
// 	// 			if path[1].Token1 == weth {
// 	// 				amount0OutSecond = new(big.Int)
// 	// 				amount1OutSecond = wethOut
// 	// 			}
// 	// 			secondTarget := common.Address(path[1].Address)

// 	// 			// fmt.Println(wethIn, [2]common.Address{firstTarget, secondTarget}, []*big.Int{amount0OutFirst, amount0OutSecond}, []*big.Int{amount1OutFirst, amount1OutSecond})

// 	// 			// run bundle
// 	// 			executor, err := web3.Eth.NewContract(config.BUNDLE_EXECTOR_ABI, config.Get().BUNDLE_EXECUTOR_ADDRESS.Hex())
// 	// 			if err != nil {
// 	// 				panic(err)
// 	// 			}
// 	// 			data, err := executor.EncodeABI("twohop", wethIn, [2]common.Address{firstTarget, secondTarget}, []*big.Int{amount0OutFirst, amount0OutSecond}, []*big.Int{amount1OutFirst, amount1OutSecond})
// 	// 			if err != nil {
// 	// 				panic(err)
// 	// 			}
// 	// 			// TODO Gas estimation
// 	// 			call := &types.CallMsg{
// 	// 				From: web3.Eth.Address(),
// 	// 				To:   executor.Address(),
// 	// 				Data: data,
// 	// 				Gas:  types.NewCallMsgBigInt(big.NewInt(types.MAX_GAS_LIMIT)),
// 	// 			}
// 	// 			gasLimit, err := web3.Eth.EstimateGas(call)
// 	// 			if err != nil {
// 	// 				sugar.Error(err)
// 	// 				// if strings.Contains(fmt.Sprint(err), "-32000") {
// 	// 				// 	panic(2)
// 	// 				// }
// 	// 			} else {
// 	// 				sugar.Info("Estimate gas limit %v\n", gasLimit)

// 	// 				// 0.006 ether
// 	// 				gasGweiPrediction := big.NewInt(5000000000000000)
// 	// 				netProfit := new(big.Int).Sub(arbProfit, gasGweiPrediction)

// 	// 				sugar.Info("Estimated Profit ", path, arbProfit, " SUB GAS ", netProfit)
// 	// 				if netProfit.Sign() == 1 {
// 	// 					// TODO not sync
// 	// 					tx, err := web3.Eth.SyncSendEIP1559RawTransaction(
// 	// 						executor.Address(),
// 	// 						new(big.Int),
// 	// 						160000,
// 	// 						gasTipCap,
// 	// 						gasFeeCap,
// 	// 						data,
// 	// 					)
// 	// 					if err != nil {
// 	// 						panic(err)
// 	// 					}
// 	// 					if err == nil {
// 	// 						sugar.Info("tx hash %v\n", tx.TxHash)
// 	// 					}
// 	// 				}
// 	// 			}
// 	// 		}
// 	// 	}
// 	// }
// }
