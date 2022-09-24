package programs

import (
	"arbitrage_go/blockchain"
	"arbitrage_go/config"
	"arbitrage_go/counter"
	"arbitrage_go/logging"
	"arbitrage_go/uniswap"
	"fmt"
	"log"
	"math/big"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/chenzhijie/go-web3"
	"github.com/chenzhijie/go-web3/eth"
	"github.com/chenzhijie/go-web3/types"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

func Arbitrage2Main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	sugar := logging.GetSugar()

	web3 := blockchain.GetWeb3()
	balanceOf := blockchain.GetWMATICBalance()
	balanceOfMu := sync.Mutex{}
	sugar.Info("WMATIC Balance for ", config.Get().BUNDLE_EXECUTOR_ADDRESS, ": ", web3.Utils.FromWei(balanceOf))

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
	executeCounter := counter.NewTSCounter(0)
	checkCounter := counter.NewTSCounter(0)

	sugar.Info("Updated ", len(allPairs), " Reserves")
	pairToReservesMu := sync.Mutex{}
	pairToReserves := uniswap.GetReservesForPairs(allPairs)
	relaventPairs := []uniswap.Pair{}
	relaventPairsMu := sync.Mutex{}
	relaventPairsMap := make(map[uniswap.Pair]bool)

	cycles := []uniswap.Cycle{}
	cyclesMu := sync.Mutex{}

	gasEstimateMu := sync.Mutex{}
	gasEstimate := blockchain.GetGasEstimate()

	go func() {
		uniswap.GetCyclesToChan(config.Get().WETH_ADDRESS, graph, 3, checkChan)
		sugar.Info("Done finding cycles")
	}()

	go func(balanceOfMu *sync.Mutex) {
		lastUpdate := time.Now()
		for {
			if time.Since(lastUpdate).Seconds() >= 2 {
				temp := blockchain.GetWMATICBalance()
				balanceOfMu.Lock()
				balanceOf = temp
				balanceOfMu.Unlock()
				sugar.Info("Updated balance: ", web3.Utils.FromWei(temp))
				lastUpdate = time.Now()
			}
		}
	}(&balanceOfMu)

	go func(gasEstimateMu *sync.Mutex) {
		lastUpdate := time.Now()
		for {
			if time.Since(lastUpdate).Seconds() >= 2 {
				newEstimate := blockchain.GetGasEstimate()
				gasEstimateMu.Lock()
				gasEstimate = newEstimate
				gasEstimateMu.Unlock()
				sugar.Info("Updated Gas Estimates, MaxFeePerGas: ", newEstimate.MaxFeePerGas, " GWEI")
				lastUpdate = time.Now()
			}
		}
	}(&gasEstimateMu)

	go func() {
		lastUpdate := time.Now()
		for {
			if time.Since(lastUpdate).Seconds() >= 2.3 {
				relaventPairsMu.Lock()
				temp := uniswap.GetReservesForPairs(relaventPairs)
				relaventPairsMu.Unlock()
				pairToReservesMu.Lock()
				for pair, reserve := range temp {
					pairToReserves[pair] = reserve
				}
				pairToReservesMu.Unlock()
				sugar.Info("Updated ", len(temp), " Relavent Reserves")
				lastUpdate = time.Now()
				cyclesMu.Lock()
				sugar.Info("Restarting Cycles")
				for _, cycle := range cycles {
					go CheckCycle(cycle, &pairToReserves, executeChan, &pairToReservesMu, checkCounter, gasEstimate, &gasEstimateMu, balanceOf, &balanceOfMu, sugar)
				}
				cyclesMu.Unlock()
			}
		}
	}()

	go func() {
		for cycle := range checkChan {
			// go CheckCycle(cycle, &pairToReserves, executeChan, &pairToReservesMu, checkCounter, gasEstimate, &gasEstimateMu, balanceOf, &balanceOfMu, sugar)
			relaventPairsMu.Lock()
			for _, pair := range cycle.Edges {
				_, exist := relaventPairsMap[pair]
				if !exist {
					relaventPairsMap[pair] = true
					relaventPairs = append(relaventPairs, pair)
				}
			}
			relaventPairsMu.Unlock()
			cyclesMu.Lock()
			cycles = append(cycles, cycle)
			cyclesMu.Unlock()
		}
	}()

	go func() {
		web3 := blockchain.GetWeb3()
		nonce, err := web3.Eth.GetNonce(web3.Eth.Address(), nil)
		if err != nil {
			sugar.Fatal(err)
		}
		nounceCounter := counter.NewTSCounter(nonce)
		for cycle := range executeChan {
			ExecuteCycle(cycle, nounceCounter, executeCounter, gasEstimate, &gasEstimateMu, balanceOf, &balanceOfMu, sugar)
		}
	}()

	go func() {
		for {
			time.Sleep(time.Second)
			cyclesMu.Lock()
			uniqueCycles := len(cycles)
			cyclesMu.Unlock()

			sugar.Info("Unique Cycles:", uniqueCycles, " | Checks:", checkCounter.TSGet(), " | Executes:", executeCounter.TSGet())
		}
	}()

}

func CheckCycle(cycle uniswap.Cycle, pairToReserves *map[uniswap.Pair]uniswap.Reserve, executeChan chan uniswap.Cycle, pairToReservesMu *sync.Mutex, checkCounter *counter.TSCounter, gasEstimate *eth.EstimateFee, gasEstimateMu *sync.Mutex, balanceOf *big.Int, balanceOfMu *sync.Mutex, sugar *zap.SugaredLogger) {
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
				gasLimit, err := newWeb3.Eth.EstimateGas(call)
				for strings.Contains(fmt.Sprint(err), "json unmarshal response body") || strings.Contains(fmt.Sprint(err), "timeout") {
					gasLimit, err = newWeb3.Eth.EstimateGas(call)
				}
				if err != nil {
					sugar.Error(err)
					sugar.Error(cycle.Tokens, cycle.Edges)
				} else {
					gasEstimateMu.Lock()
					maxGasWei := new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasEstimate.MaxFeePerGas)
					netProfit := new(big.Int).Sub(arbProfit, maxGasWei)
					gasEstimateMu.Unlock()
					// sugar.Info(arbProfit, netProfit)

					if netProfit.Sign() == 1 {
						executeChan <- cycle
					}
				}
			}
		}
	}
}

func ExecuteCycle(cycle uniswap.Cycle, nonceCounter *counter.TSCounter, executeCounter *counter.TSCounter, gasEstimate *eth.EstimateFee, gasEstimateMu *sync.Mutex, balanceOf *big.Int, balanceOfMu *sync.Mutex, sugar *zap.SugaredLogger) {
	pairToReserves := uniswap.GetReservesForPairs(cycle.Edges)

	// sugar := logging.GetSugar()

	newWeb3, err := web3.NewWeb3(config.Get().RPC_URL_HTTP)
	if err != nil {
		log.Fatal(err)
	}
	newWeb3.Eth.SetChainId(config.Get().CHAIN_ID)
	newWeb3.Eth.SetAccount(config.Get().PRIVATE_KEY)

	E0, E1 := uniswap.GetE0E1ForCycle(cycle, pairToReserves)

	if new(big.Int).Sub(E0, E1).Sign() == -1 {
		amountIn := uniswap.GetOptimalAmountIn(E0, E1)
		balanceOfMu.Lock()
		if big.NewInt(0).Sub(amountIn, balanceOf).Sign() == 1 {
			amountIn = balanceOf
		}
		balanceOfMu.Unlock()

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
				for strings.Contains(fmt.Sprint(err), "json unmarshal response body") || strings.Contains(fmt.Sprint(err), "timeout") {
					gasLimit, err = newWeb3.Eth.EstimateGas(call)
				}

				if err != nil {
					sugar.Error(err)
				} else {
					gasEstimateMu.Lock()
					maxGasWei := new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasEstimate.MaxFeePerGas)
					netProfit := new(big.Int).Sub(arbProfit, maxGasWei)
					gasEstimateMu.Unlock()

					fmt.Println(maxGasWei)

					if netProfit.Sign() == 1 {
						// sugar.Info("Estimated Profit ", cycle.Edges, arbProfit, " SUB GAS ", netProfit)
						gasTipCap := gasEstimate.MaxPriorityFeePerGas
						gasFeeCap := gasEstimate.MaxFeePerGas

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
							executeCounter.TSInc()
							sugar.Info("tx hash: ", hash)
						}
					}
				}
			}
		}
	}
}
