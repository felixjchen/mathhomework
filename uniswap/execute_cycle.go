package uniswap

import (
	"arbitrage_go/apps/flashbots"
	"arbitrage_go/blockchain"
	"arbitrage_go/config"
	"arbitrage_go/counter"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/chenzhijie/go-web3/eth"
	"github.com/chenzhijie/go-web3/types"
	eTypes "github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

func ExecuteCycle(cycle Cycle, nonceCounter *counter.TSCounter, executeCounter *counter.TSCounter, gasEstimate *eth.EstimateFee, gasEstimateMu *sync.Mutex, balanceOf *big.Int, balanceOfMu *sync.Mutex, sugar *zap.SugaredLogger) {
	pairToReserves := GetReservesForPairs(cycle.Edges)

	E0, E1 := GetE0E1ForCycle(cycle, pairToReserves)

	if new(big.Int).Sub(E0, E1).Sign() == -1 {
		amountIn := GetOptimalAmountIn(E0, E1)

		balanceOfMu.Lock()
		if big.NewInt(0).Sub(amountIn, balanceOf).Sign() == 1 {
			sugar.Info("MIN ", amountIn, balanceOf)
			amountIn = balanceOf
		}
		balanceOfMu.Unlock()
		// sugar.Info("Estimated Profit ", newWeb3.Utils.FromWei(amountIn), newWeb3.Utils.FromWei(arbProfit))

		if amountIn.Sign() == 1 {
			amountsOut := GetAmountsOutCycle(pairToReserves, amountIn, cycle)
			arbProfit := big.NewInt(0).Sub(amountsOut[len(amountsOut)-1], amountIn)

			if arbProfit.Sign() == 1 {
				targets := GetCycleTargets(cycle)
				cycleAmountsOut := GetCycleAmountsOut(cycle, amountsOut)

				newWeb3 := blockchain.GetWeb3()
				// run bundle
				executor, err := newWeb3.Eth.NewContract(config.BUNDLE_EXECTOR_ABI, config.Get().BUNDLE_EXECUTOR_ADDRESS.Hex())
				if err != nil {
					panic(err)
				}
				data := GetPayload(cycle, executor, amountIn, targets, cycleAmountsOut)
				call := &types.CallMsg{
					From: newWeb3.Eth.Address(),
					To:   executor.Address(),
					Data: data,
					Gas:  types.NewCallMsgBigInt(big.NewInt(types.MAX_GAS_LIMIT)),
				}

				gasLimit, err := newWeb3.Eth.EstimateGas(call)

				if err != nil {
					sugar.Error("ERROR IN EXECUTE Q: ", newWeb3.Utils.FromWei(arbProfit))
					sugar.Error(err)
				} else {
					gasEstimateMu.Lock()
					maxGasWei := new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasEstimate.MaxFeePerGas)
					netProfit := new(big.Int).Sub(arbProfit, maxGasWei)
					gasEstimateMu.Unlock()

					if netProfit.Sign() == 1 {
						sugar.Info("Estimated Profit ", cycle.Edges, newWeb3.Utils.FromWei(arbProfit), " SUB GAS ", newWeb3.Utils.FromWei(netProfit))
						gasTipCap := gasEstimate.MaxPriorityFeePerGas
						gasFeeCap := gasEstimate.MaxFeePerGas

						nonceCounter.Lock()
						defer nonceCounter.Unlock()
						nonce := nonceCounter.Get()

						tx, err := newWeb3.Eth.NewEIP1559Tx(
							executor.Address(),
							new(big.Int),
							gasLimit,
							gasTipCap,
							gasFeeCap,
							data,
							nonce,
						)
						if err != nil {
							sugar.Fatal(err)
						}
						bundleTxs := make([]*eTypes.Transaction, 0)
						bundleTxs = append(bundleTxs, tx)

						fb, err := flashbots.NewFlashBot(flashbots.TestRelayURL, config.Get().PRIVATE_KEY)
						if err != nil {
							sugar.Fatal(err)
						}
						currentBlockNumber, err := newWeb3.Eth.GetBlockNumber()
						if err != nil {
							sugar.Fatal(1, err)
						}
						resp, err := fb.Simulate(
							bundleTxs,
							big.NewInt(int64(currentBlockNumber)),
							"latest",
						)
						if strings.Contains(fmt.Sprint(err), "nonce too low") {
							nonceCounter.Inc()
							executeCounter.TSInc()
							return
						}
						sugar.Info("Resp %s", resp)
						targetBlockNumber := big.NewInt(int64(currentBlockNumber) + 1)
						bundleResp, err := fb.SendBundle(
							bundleTxs,
							targetBlockNumber,
						)
						if err != nil {
							sugar.Fatal(2, err)
						}
						sugar.Info("bundle resp %v\n", bundleResp)

						stat, err := fb.GetBundleStats(bundleResp.BundleHash, targetBlockNumber)
						if err != nil {
							sugar.Fatal(4, err)
						}
						sugar.Info("bundle stat %v\n", stat)

						// hash, err := newWeb3.Eth.SendRawEIP1559TransactionWithNonce(
						// 	nonce,
						// 	executor.Address(),
						// 	new(big.Int),
						// 	gasLimit,
						// 	gasTipCap,
						// 	gasFeeCap,
						// 	data,
						// )
						// if err != nil {
						// 	panic(err)
						// } else {
						// 	nonceCounter.Inc()
						// 	executeCounter.TSInc()
						// 	sugar.Info("tx hash: ", hash)
						// 	panic(1)
						// }
					}
				}
			}
		}
	}
}
