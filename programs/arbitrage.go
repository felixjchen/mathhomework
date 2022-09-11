package programs

import (
	"arbitrage_go/blockchain"
	"arbitrage_go/config"
	"arbitrage_go/uniswap"
	"arbitrage_go/util"
	"fmt"
	"log"
	"math/big"

	"github.com/chenzhijie/go-web3/types"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

func Arbitrage() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	sugar := logger.Sugar()

	web3 := blockchain.GetWeb3()

	allPools := uniswap.GetAllPools()
	allPools = uniswap.FilterPools(util.TokenBlacklistFilter, allPools)
	sugar.Info("Got ", len(allPools), " pools")

	// pathing
	// adjacency list
	tokensToPools := make(map[common.Address][]uniswap.Pool)
	for _, pool := range allPools {
		tokensToPools[pool.Token0] = append(tokensToPools[pool.Token0], pool)
		tokensToPools[pool.Token1] = append(tokensToPools[pool.Token1], pool)
	}
	sugar.Info("Created Graph")

	// pool has weth
	wethPools := uniswap.FilterPools(util.WethFilter, allPools)

	// two hop
	weth := config.Get().WETH_ADDRESS
	pathes := uniswap.GetTwoHops(tokensToPools)
	sugar.Info("Found all 2-hops")

	// TODO only if needed
	poolToReserves := uniswap.UpdateReservesForPools(wethPools)
	sugar.Info("Updated Reserves")

	// Simulate path
	for _, path := range pathes {

		intermediateToken := util.Ternary(path[0].Token0 != weth, path[0].Token0, path[0].Token1)

		R0 := util.Ternary(path[0].Token0 == weth, poolToReserves[path[0]].Reserve0, poolToReserves[path[0]].Reserve1)
		R1 := util.Ternary(path[0].Token0 == intermediateToken, poolToReserves[path[0]].Reserve0, poolToReserves[path[0]].Reserve1)

		R1_ := util.Ternary(path[1].Token0 == intermediateToken, poolToReserves[path[1]].Reserve0, poolToReserves[path[1]].Reserve1)
		R2 := util.Ternary(path[1].Token0 == weth, poolToReserves[path[1]].Reserve0, poolToReserves[path[1]].Reserve1)

		E0, E1 := uniswap.GetE0E1(R0, R1, R1_, R2)

		if big.NewInt(0).Sub(E0, E1).Sign() == -1 {

			wethIn := uniswap.GetOptimalWethIn(E0, E1)

			// Min on current balance
			if big.NewInt(0).Sub(wethIn, big.NewInt(20000000000000000)).Sign() == 1 {
				wethIn = big.NewInt(20000000000000000)
			}

			if wethIn.Sign() == 1 {
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
				arbProfitSubGas := big.NewInt(0).Sub(arbProfit, big.NewInt(5971360001492840))

				if arbProfitSubGas.Sign() == 1 {
					sugar.Info(" PROFIT ", path, arbProfit, " SUB GAS ", arbProfitSubGas)
					// Build first txn
					amount0OutFirst := big.NewInt(0)
					amount1OutFirst := intermediateAmount
					if path[0].Token1 == weth {
						amount0OutFirst = intermediateAmount
						amount1OutFirst = big.NewInt(0)
					}
					firstTarget := common.Address(path[0].Address)

					// build second txn
					amount0OutSecond := wethOut
					amount1OutSecond := big.NewInt(0)
					if path[1].Token1 == weth {
						amount0OutSecond = big.NewInt(0)
						amount1OutSecond = wethOut
					}
					secondTarget := common.Address(path[1].Address)

					fmt.Println(wethIn, [2]common.Address{firstTarget, secondTarget}, []*big.Int{amount0OutFirst, amount0OutSecond}, []*big.Int{amount1OutFirst, amount1OutSecond})

					// run bundle
					executor, err := web3.Eth.NewContract(config.BUNDLE_EXECTOR_ABI, config.Get().BUNDLE_EXECUTOR_ADDRESS.Hex())
					if err != nil {
						panic(err)
					}
					data, err := executor.EncodeABI("twohop", wethIn, [2]common.Address{firstTarget, secondTarget}, []*big.Int{amount0OutFirst, amount0OutSecond}, []*big.Int{amount1OutFirst, amount1OutSecond})
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
						// panic(err)
						sugar.Error(err)
					} else {
						sugar.Info("Estimate gas limit %v\n", gasLimit)

						// TODO not sync
						tx, err := web3.Eth.SyncSendEIP1559RawTransaction(
							executor.Address(),
							big.NewInt(0),
							gasLimit*2,
							web3.Utils.ToGWei(40),
							web3.Utils.ToGWei(325),
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