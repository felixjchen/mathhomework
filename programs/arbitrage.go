package programs

import (
	"arbitrage_go/blockchain"
	"arbitrage_go/config"
	"arbitrage_go/counter"
	"arbitrage_go/database"
	"arbitrage_go/logging"
	"arbitrage_go/uniswap"
	"arbitrage_go/util"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/chenzhijie/go-web3"
	"github.com/chenzhijie/go-web3/eth"
	"github.com/chenzhijie/go-web3/types"
	"go.uber.org/zap"
)

const MAX_CHECK_SIZE = 10000
const ESTIMATE_TIMEOUT = 2

// const MAX_QUERY_SIZE = 100000

// 0.002 ETHER
const BATCH_THRESHOLD = 2000000000000000

type RoughHopGasLimit struct {
	RoughGasLimit uint64
	LastTimestamp time.Time
}

func ArbitrageMain() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	sugar := logging.GetSugar("arb")

	web3 := blockchain.GetWeb3()

	nonce, err := web3.Eth.GetNonce(web3.Eth.Address(), nil)
	if err != nil {
		sugar.Fatal(err)
	}
	nounceCounter := counter.NewTSCounter(nonce)

	executeChan := make(chan uniswap.Cycle)
	batchChan := make(chan uniswap.BatchCandidate)
	executeCounter := counter.NewTSCounter(0)
	checkCounter := counter.NewTSCounter(0)

	database := database.NewDBConn(sugar)
	cycleHashes := database.GetCycleHashes()
	sugar.Info("Found ", len(cycleHashes), " cycles")
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(cycleHashes), func(i, j int) { cycleHashes[i], cycleHashes[j] = cycleHashes[j], cycleHashes[i] })

	relaventPairs := database.GetPairs()
	sugar.Info("Found ", len(relaventPairs), " relavent pairs")

	// Get all reserves to start with
	pairToReservesMu := sync.Mutex{}
	pairToReserves := uniswap.GetReservesForPairs(relaventPairs)
	sugar.Info("Updated ", len(relaventPairs), " Reserves")

	balanceOfMu := sync.Mutex{}
	balanceOf := blockchain.GetWMATICBalance()
	go func(balanceOfMu *sync.Mutex) {
		lastUpdate := time.Now()
		for {
			if time.Since(lastUpdate).Seconds() >= 2 {
				temp := blockchain.GetWMATICBalance()
				balanceOfMu.Lock()
				balanceOf = temp
				sugar.Info("Updated balance: ", web3.Utils.FromWei(balanceOf))
				balanceOfMu.Unlock()
				lastUpdate = time.Now()
			}
		}
	}(&balanceOfMu)

	gasEstimateMu := sync.Mutex{}
	gasEstimate := blockchain.GetGasEstimate()
	go func(gasEstimateMu *sync.Mutex) {
		lastUpdate := time.Now()
		for {
			if time.Since(lastUpdate).Seconds() >= 2 {
				newEstimate := blockchain.GetGasEstimate()
				gasEstimateMu.Lock()
				sugar.Info("MaxFeePerGas: ", gasEstimate.MaxFeePerGas, " GWEI")
				gasEstimate = newEstimate
				gasEstimateMu.Unlock()
				lastUpdate = time.Now()
			}
		}
	}(&gasEstimateMu)

	roughHopGasLimitMu := sync.Mutex{}
	roughHopGasLimit := make(map[int]RoughHopGasLimit)
	for i := 0; i < 10; i++ {
		roughHopGasLimit[i] = RoughHopGasLimit{0, time.Now()}
	}

	go func() {
		lastUpdate := time.Now()
		for {
			if time.Since(lastUpdate).Seconds() >= 2 {
				temp := uniswap.GetReservesForPairs(relaventPairs)
				pairToReservesMu.Lock()
				for pair, reserve := range temp {
					pairToReserves[pair] = reserve
				}
				pairToReservesMu.Unlock()
				sugar.Info("Updated: ", len(relaventPairs), " relaventPairs reserves")
				lastUpdate = time.Now()
			}
		}
	}()

	go func() {
		for {
			for i := 0; i < len(cycleHashes); i += MAX_CHECK_SIZE {
				j := util.Ternary(i+MAX_CHECK_SIZE < len(cycleHashes), i+MAX_CHECK_SIZE, len(cycleHashes)-1)
				cycles := database.GetCycles(cycleHashes[i:j])
				wg := sync.WaitGroup{}
				for _, cycle := range cycles {
					wg.Add(1)
					go CheckCycleWG(cycle, &pairToReserves, executeChan, &pairToReservesMu, checkCounter, gasEstimate, &gasEstimateMu, balanceOf, &balanceOfMu, &roughHopGasLimitMu, &roughHopGasLimit, sugar, &wg)
				}
				wg.Wait()
			}
			sugar.Info("Done Cycles")
		}
	}()

	go func() {
		for cycle := range executeChan {
			uniswap.ExecuteCycle(cycle, nounceCounter, executeCounter, gasEstimate, &gasEstimateMu, balanceOf, &balanceOfMu, sugar)
		}
	}()

	go uniswap.StartBatchQueue(sugar, batchChan, nounceCounter, executeCounter, gasEstimate, &gasEstimateMu, balanceOf, &balanceOfMu)

	go func() {
		for {
			time.Sleep(time.Second)
			sugar.Info("Checks:", checkCounter.TSGet(), " | Executes:", executeCounter.TSGet())
		}
	}()
}

func CheckCycleWG(cycle uniswap.Cycle, pairToReserves *map[uniswap.Pair]uniswap.Reserve, executeChan chan uniswap.Cycle, pairToReservesMu *sync.Mutex, checkCounter *counter.TSCounter, gasEstimate *eth.EstimateFee, gasEstimateMu *sync.Mutex, balanceOf *big.Int, balanceOfMu *sync.Mutex, roughHopGasLimitMu *sync.Mutex, roughHopGasLimit *map[int]RoughHopGasLimit, sugar *zap.SugaredLogger, wg *sync.WaitGroup) {
	defer checkCounter.TSInc()

	newWeb3, err := web3.NewWeb3(config.Get().RPC_URL_HTTP)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	newWeb3.Eth.SetChainId(config.Get().CHAIN_ID)
	newWeb3.Eth.SetAccount(config.Get().PRIVATE_KEY)

	pairToReservesMu.Lock()
	E0, E1 := uniswap.GetE0E1ForCycle(cycle, *pairToReserves)
	pairToReservesMu.Unlock()

	if new(big.Int).Sub(E0, E1).Sign() == -1 {
		amountIn := uniswap.GetOptimalAmountIn(E0, E1)

		// Min on current balance
		balanceOfMu.Lock()
		if big.NewInt(0).Sub(amountIn, balanceOf).Sign() == 1 {
			sugar.Info("MIN ", amountIn, balanceOf)
			amountIn = balanceOf
		}
		balanceOfMu.Unlock()
		if amountIn.Sign() == 1 {
			pairToReservesMu.Lock()
			amountsOut := uniswap.GetAmountsOutCycle(*pairToReserves, amountIn, cycle)
			pairToReservesMu.Unlock()
			arbProfit := big.NewInt(0).Sub(amountsOut[len(amountsOut)-1], amountIn)

			if arbProfit.Sign() == 1 {
				targets := uniswap.GetCycleTargets(cycle)
				cycleAmountsOut := uniswap.GetCycleAmountsOut(cycle, amountsOut)

				// run bundle
				executor, err := newWeb3.Eth.NewContract(config.BUNDLE_EXECTOR_ABI, config.Get().BUNDLE_EXECUTOR_ADDRESS.Hex())
				if err != nil {
					panic(err)
				}
				hopLen := len(cycle.Edges)

				roughHopGasLimitMu.Lock()
				noTimestamp := (*roughHopGasLimit)[hopLen].RoughGasLimit == 0
				needsRefresh := time.Since((*roughHopGasLimit)[hopLen].LastTimestamp).Seconds() > ESTIMATE_TIMEOUT

				if noTimestamp || needsRefresh {
					data := uniswap.GetPayload(cycle, executor, amountIn, targets, cycleAmountsOut)
					call := &types.CallMsg{
						From: newWeb3.Eth.Address(),
						To:   executor.Address(),
						Data: data,
						Gas:  types.NewCallMsgBigInt(big.NewInt(types.MAX_GAS_LIMIT)),
					}
					gasLimit, err := newWeb3.Eth.EstimateGas(call)
					for strings.Contains(fmt.Sprint(err), "json unmarshal response body") || strings.Contains(fmt.Sprint(err), "timeout") {
						gasLimit, err = newWeb3.Eth.EstimateGas(call)
					}

					/// sometimes shit
					if gasLimit != 0 {
						entry := (*roughHopGasLimit)[hopLen]
						entry.RoughGasLimit = gasLimit
						entry.LastTimestamp = time.Now()
						(*roughHopGasLimit)[hopLen] = entry
					}
				}

				gasLimit := (*roughHopGasLimit)[hopLen].RoughGasLimit
				roughHopGasLimitMu.Unlock()

				if err != nil {
					sugar.Error(err)
					// sugar.Error(cycle.Tokens, cycle.Edges)
				} else {
					gasEstimateMu.Lock()
					maxGasWei := new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasEstimate.MaxFeePerGas)
					netProfit := new(big.Int).Sub(arbProfit, maxGasWei)
					gasEstimateMu.Unlock()
					// sugar.Info(arbProfit, netProfit)
					// if new(big.Int).Add(netProfit, big.NewInt(BATCH_THRESHOLD)).Sign() >= 1 {
					// 	sugar.Info("BATCH CANDIDATE ", netProfit)
					// }

					if netProfit.Sign() == 1 {
						executeChan <- cycle
					}
				}
			}
		}
	}
	wg.Done()
}

// func ExecuteCycle(cycle uniswap.Cycle, nonceCounter *counter.TSCounter, executeCounter *counter.TSCounter, gasEstimate *eth.EstimateFee, gasEstimateMu *sync.Mutex, balanceOf *big.Int, balanceOfMu *sync.Mutex, sugar *zap.SugaredLogger) {
// 	pairToReserves := uniswap.GetReservesForPairs(cycle.Edges)

// 	// sugar := logging.GetSugar()

// 	newWeb3, err := web3.NewWeb3(config.Get().RPC_URL_HTTP)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	newWeb3.Eth.SetChainId(config.Get().CHAIN_ID)
// 	newWeb3.Eth.SetAccount(config.Get().PRIVATE_KEY)

// 	E0, E1 := uniswap.GetE0E1ForCycle(cycle, pairToReserves)

// 	if new(big.Int).Sub(E0, E1).Sign() == -1 {
// 		amountIn := uniswap.GetOptimalAmountIn(E0, E1)
// 		balanceOfMu.Lock()
// 		if big.NewInt(0).Sub(amountIn, balanceOf).Sign() == 1 {
// 			sugar.Info("MIN", amountIn, balanceOf)
// 			amountIn = balanceOf
// 		}
// 		balanceOfMu.Unlock()

// 		if amountIn.Sign() == 1 {
// 			amountsOut := uniswap.GetAmountsOutCycle(pairToReserves, amountIn, cycle)
// 			arbProfit := big.NewInt(0).Sub(amountsOut[len(amountsOut)-1], amountIn)

// 			if arbProfit.Sign() == 1 {
// 				targets := uniswap.GetCycleTargets(cycle)
// 				cycleAmountsOut := uniswap.GetCycleAmountsOut(cycle, amountsOut)

// 				// run bundle
// 				executor, err := newWeb3.Eth.NewContract(config.BUNDLE_EXECTOR_ABI, config.Get().BUNDLE_EXECUTOR_ADDRESS.Hex())
// 				if err != nil {
// 					panic(err)
// 				}

// 				data, err := executor.EncodeABI("hoppity", amountIn, targets, cycleAmountsOut)
// 				if err != nil {
// 					panic(err)
// 				}
// 				// TODO Gas estimation
// 				call := &types.CallMsg{
// 					From: newWeb3.Eth.Address(),
// 					To:   executor.Address(),
// 					Data: data,
// 					Gas:  types.NewCallMsgBigInt(big.NewInt(types.MAX_GAS_LIMIT)),
// 				}
// 				gasLimit, err := newWeb3.Eth.EstimateGas(call)
// 				for strings.Contains(fmt.Sprint(err), "json unmarshal response body") || strings.Contains(fmt.Sprint(err), "timeout") {
// 					gasLimit, err = newWeb3.Eth.EstimateGas(call)
// 				}

// 				if err != nil {
// 					sugar.Error("ERROR IN EXECUTE Q")
// 					sugar.Error(err)
// 				} else {
// 					gasEstimateMu.Lock()
// 					maxGasWei := new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasEstimate.MaxFeePerGas)
// 					netProfit := new(big.Int).Sub(arbProfit, maxGasWei)
// 					gasEstimateMu.Unlock()

// 					fmt.Println(maxGasWei)

// 					if netProfit.Sign() == 1 {
// 						// sugar.Info("Estimated Profit ", cycle.Edges, arbProfit, " SUB GAS ", netProfit)
// 						gasTipCap := gasEstimate.MaxPriorityFeePerGas
// 						gasFeeCap := gasEstimate.MaxFeePerGas

// 						nonceCounter.Lock()
// 						defer nonceCounter.Unlock()
// 						nonce := nonceCounter.Get()
// 						hash, err := newWeb3.Eth.SendRawEIP1559TransactionWithNonce(
// 							nonce,
// 							executor.Address(),
// 							new(big.Int),
// 							gasLimit,
// 							gasTipCap,
// 							gasFeeCap,
// 							data,
// 						)
// 						if err != nil {
// 							fmt.Println("PANIC", err)
// 							// panic(err)
// 						} else {
// 							nonceCounter.Inc()
// 							executeCounter.TSInc()
// 							sugar.Info("tx hash: ", hash)
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}
// }
