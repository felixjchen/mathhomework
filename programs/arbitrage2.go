package programs

import (
	"arbitrage_go/blockchain"
	"arbitrage_go/config"
	"arbitrage_go/uniswap"
	"fmt"
	"log"
	"math/big"
	"runtime"
	"sync"
	"time"

	"github.com/chenzhijie/go-web3"
	"github.com/chenzhijie/go-web3/types"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

type TSCounter struct {
	count uint64
	mu    *sync.Mutex
}

func NewTSCounter(i uint64) *TSCounter {
	return &TSCounter{
		count: i,
		mu:    &sync.Mutex{},
	}
}

func (n *TSCounter) Get() uint64 {
	return n.count
}

func (n *TSCounter) Inc() {
	n.count++
}

func (n *TSCounter) Dec() {
	n.count--
}

func (n *TSCounter) TSGet() uint64 {
	n.Lock()
	defer n.Unlock()
	return n.count
}

func (n *TSCounter) TSInc() {
	n.Lock()
	defer n.Unlock()
	n.count++
}

func (n *TSCounter) TSDec() {
	n.Lock()
	defer n.Unlock()
	n.count--
}

func (n *TSCounter) Lock() {
	n.mu.Lock()
}
func (n *TSCounter) Unlock() {
	n.mu.Unlock()
}

func Arbitrage2Main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

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

	executeChan := make(chan uniswap.Cycle)
	checkChan := make(chan uniswap.Cycle)

	go func() {
		for {
			uniswap.GetCycles(config.Get().WETH_ADDRESS, graph, 3, checkChan)
			sugar.Info("RESTARTING CYCLES")
		}
	}()

	pairToReserves := uniswap.GetReservesForPairs(allPairs)
	sugar.Info("Updated ", len(allPairs), " Reserves")
	mu := sync.Mutex{}
	go func() {
		lastUpdate := time.Now()
		for {
			if time.Since(lastUpdate).Seconds() >= 2.3 {
				temp := uniswap.GetReservesForPairs(allPairs)
				mu.Lock()
				pairToReserves = temp
				mu.Unlock()
				sugar.Info("Updated ", len(allPairs), " Reserves Live")
			}
		}
	}()

	go func() {
		for cycle := range checkChan {
			go CheckCycle(cycle, &pairToReserves, executeChan, &mu)
		}
	}()

	go func() {
		web3 := blockchain.GetWeb3()
		nonce, err := web3.Eth.GetNonce(web3.Eth.Address(), nil)
		if err != nil {
			log.Fatal(err)
		}
		nounceCounter := NewTSCounter(nonce)
		for cycle := range executeChan {
			ExecuteCycle(cycle, nounceCounter)
		}
	}()

	// go func(liveChecks *int, liveExecutes *int) {
	// 	for {
	// 		time.Sleep(5 * time.Second)
	// 		fmt.Println(*liveChecks, *liveExecutes)
	// 	}
	// }(&liveChecks, &liveExecutes)

	// go func() {
	// 	for {
	// 		time.Sleep(3 * time.Second)
	// 		sugar.Info("Check Q:", len(checkChan), " Execute Q:", len(executeChan))
	// 	}
	// }()

}

func CheckCycle(cycle uniswap.Cycle, pairToReserves *map[uniswap.Pair]uniswap.Reserve, executeChan chan uniswap.Cycle, mu *sync.Mutex) {
	newWeb3, err := web3.NewWeb3(config.Get().RPC_URL_HTTP)
	if err != nil {
		log.Fatal(err)
	}
	newWeb3.Eth.SetChainId(config.Get().CHAIN_ID)
	newWeb3.Eth.SetAccount(config.Get().PRIVATE_KEY)

	mu.Lock()
	E0, E1 := uniswap.GetE0E1ForCycle(cycle, *pairToReserves)
	mu.Unlock()

	if new(big.Int).Sub(E0, E1).Sign() == -1 {
		amountIn := uniswap.GetOptimalAmountIn(E0, E1)

		// Min on current balance
		if config.PROD {
			if big.NewInt(0).Sub(amountIn, big.NewInt(960000000000000000)).Sign() == 1 {
				amountIn = big.NewInt(960000000000000000)
			}
		}

		if amountIn.Sign() == 1 {
			mu.Lock()
			amountsOut := uniswap.GetAmountsOutCycle(*pairToReserves, amountIn, cycle)
			mu.Unlock()
			arbProfit := big.NewInt(0).Sub(amountsOut[len(amountsOut)-1], amountIn)

			if arbProfit.Sign() == 1 {
				targets := uniswap.GetCycleTargets(cycle)
				cycleAmountsOut := uniswap.GetCycleAmountsOut(cycle, amountsOut)

				// run bundle
				executor, err := newWeb3.Eth.NewContract(config.BUNDLE_EXECTOR_ABI, config.Get().BUNDLE_EXECUTOR_ADDRESS.Hex())
				if err != nil {
					panic(err)
				}

				data, err := executor.EncodeABI("hoppity", amountIn, targets, cycleAmountsOut)
				if err != nil {
					panic(err)
				}
				call := &types.CallMsg{
					From: newWeb3.Eth.Address(),
					To:   executor.Address(),
					Data: data,
					Gas:  types.NewCallMsgBigInt(big.NewInt(types.MAX_GAS_LIMIT)),
				}
				_, err = newWeb3.Eth.EstimateGas(call)
				if err == nil {
					// 0.01 ether
					gasGweiPrediction := big.NewInt(10000000000000000)
					netProfit := new(big.Int).Sub(arbProfit, gasGweiPrediction)
					fmt.Println("Estimated Profit ", cycle, arbProfit, " SUB GAS ", netProfit)

					if netProfit.Sign() == 1 {
						executeChan <- cycle
					}
				}
			}
		}
	}

}

func ExecuteCycle(cycle uniswap.Cycle, nonceCounter *TSCounter) {
	// fmt.Println(cycle)
	pairToReserves := uniswap.GetReservesForPairs(cycle.Edges)

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	sugar := logger.Sugar()

	newWeb3, err := web3.NewWeb3(config.Get().RPC_URL_HTTP)
	if err != nil {
		log.Fatal(err)
	}
	newWeb3.Eth.SetChainId(config.Get().CHAIN_ID)
	newWeb3.Eth.SetAccount(config.Get().PRIVATE_KEY)

	E0, E1 := uniswap.GetE0E1ForCycle(cycle, pairToReserves)

	if new(big.Int).Sub(E0, E1).Sign() == -1 {
		amountIn := uniswap.GetOptimalAmountIn(E0, E1)
		// Min on current balance
		if big.NewInt(0).Sub(amountIn, big.NewInt(960000000000000000)).Sign() == 1 {
			amountIn = big.NewInt(960000000000000000)
		}

		if amountIn.Sign() == 1 {
			amountsOut := uniswap.GetAmountsOutCycle(pairToReserves, amountIn, cycle)
			arbProfit := big.NewInt(0).Sub(amountsOut[len(amountsOut)-1], amountIn)

			if arbProfit.Sign() == 1 {
				targets := uniswap.GetCycleTargets(cycle)
				cycleAmountsOut := uniswap.GetCycleAmountsOut(cycle, amountsOut)

				// run bundle
				executor, err := newWeb3.Eth.NewContract(config.BUNDLE_EXECTOR_ABI, config.Get().BUNDLE_EXECUTOR_ADDRESS.Hex())
				if err != nil {
					panic(err)
				}

				data, err := executor.EncodeABI("hoppity", amountIn, targets, cycleAmountsOut)
				if err != nil {
					panic(err)
				}
				// TODO Gas estimation
				call := &types.CallMsg{
					From: newWeb3.Eth.Address(),
					To:   executor.Address(),
					Data: data,
					Gas:  types.NewCallMsgBigInt(big.NewInt(types.MAX_GAS_LIMIT)),
				}
				gasLimit, err := newWeb3.Eth.EstimateGas(call)
				if err != nil {
					// sugar.Error(err)
				} else {
					// 0.01 ether
					gasGweiPrediction := big.NewInt(10000000000000000)
					netProfit := new(big.Int).Sub(arbProfit, gasGweiPrediction)

					if netProfit.Sign() == 1 {
						// sugar.Info("Estimate gas limit %v\n", gasLimit)
						// sugar.Info("Estimated Profit ", cycle.Edges, arbProfit, " SUB GAS ", netProfit)
						gasTipCap := newWeb3.Utils.ToGWei(38)
						gasFeeCap := newWeb3.Utils.ToGWei(40)
						if !config.PROD {
							gasTipCap = newWeb3.Utils.ToGWei(3.4)
							gasFeeCap = newWeb3.Utils.ToGWei(3.5)
						}

						nonceCounter.Lock()
						defer nonceCounter.Unlock()
						nonce := nonceCounter.Get()
						hash, err := newWeb3.Eth.SendRawEIP1559TransactionWithNonce(
							nonce,
							executor.Address(),
							new(big.Int),
							gasLimit,
							gasTipCap,
							gasFeeCap,
							data,
						)
						if err != nil {
							fmt.Println("PANIC", err)
							// panic(err)
						} else {
							nonceCounter.Inc()
							sugar.Info("tx hash: ", hash)
							// panic(1)
						}
					}
				}
			}
		}
	}
}
