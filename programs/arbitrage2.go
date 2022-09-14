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
					cycleAmountsOut := uniswap.GetCycleAmountsOut(cycle, amountsOut)

					// run bundle
					executor, err := web3.Eth.NewContract(config.BUNDLE_EXECTOR_ABI, config.Get().BUNDLE_EXECUTOR_ADDRESS.Hex())
					if err != nil {
						panic(err)
					}
					// fmt.Println(cycle, amountsOut)
					// fmt.Println(amountIn, targets, amounts0Out, amounts1Out)

					data, err := executor.EncodeABI("hoppity", amountIn, targets, cycleAmountsOut)
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
