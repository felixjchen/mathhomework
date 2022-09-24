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
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/chenzhijie/go-web3"
	"github.com/chenzhijie/go-web3/eth"
	"github.com/chenzhijie/go-web3/types"
	"go.uber.org/zap"
)

const MAX_CHECK_SIZE = 25
const MAX_QUERY_SIZE = 100000

func Arbitrage3Main() {
	web3 := blockchain.GetWeb3()
	runtime.GOMAXPROCS(runtime.NumCPU())

	sugar := logging.GetSugar()

	executeChan := make(chan uniswap.Cycle)
	executeCounter := counter.NewTSCounter(0)
	checkCounter := counter.NewTSCounter(0)

	database := database.NewDBConn()
	cycleHashes := database.GetCycleHashes()

	relaventPairs := []uniswap.Pair{}
	relaventPairsMap := make(map[uniswap.Pair]bool)
	for i := 0; i < len(cycleHashes); i += MAX_QUERY_SIZE {
		j := util.Ternary(i+MAX_QUERY_SIZE < len(cycleHashes), i+MAX_QUERY_SIZE, len(cycleHashes)-1)
		cycles := database.GetCycles(cycleHashes[i:j])
		for _, cycle := range cycles {
			for _, pair := range cycle.Edges {
				_, exist := relaventPairsMap[pair]
				if !exist {
					relaventPairsMap[pair] = true
					relaventPairs = append(relaventPairs, pair)
				}
			}
		}
	}
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
				gasEstimate = newEstimate
				gasEstimateMu.Unlock()
				lastUpdate = time.Now()
			}
		}
	}(&gasEstimateMu)

	go func() {
		lastUpdate := time.Now()
		for {
			if time.Since(lastUpdate).Seconds() >= 2.3 {
				temp := uniswap.GetReservesForPairs(relaventPairs)
				pairToReservesMu.Lock()
				for pair, reserve := range temp {
					pairToReserves[pair] = reserve
				}
				pairToReservesMu.Unlock()
				lastUpdate = time.Now()
			}
		}
	}()

	go func() {
		for i := 0; i < len(cycleHashes); i += MAX_CHECK_SIZE {
			j := util.Ternary(i+MAX_CHECK_SIZE < len(cycleHashes), i+MAX_CHECK_SIZE, len(cycleHashes)-1)
			cycles := database.GetCycles(cycleHashes[i:j])
			wg := sync.WaitGroup{}
			for _, cycle := range cycles {
				wg.Add(1)
				go CheckCycleWG(cycle, &pairToReserves, executeChan, &pairToReservesMu, checkCounter, gasEstimate, &gasEstimateMu, balanceOf, &balanceOfMu, sugar, &wg)
			}
			wg.Wait()
		}
	}()

	go func() {
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
			sugar.Info("Checks:", checkCounter.TSGet(), " | Executes:", executeCounter.TSGet())
			balanceOfMu.Lock()
			sugar.Info("Updated balance: ", web3.Utils.FromWei(balanceOf))
			balanceOfMu.Unlock()
			gasEstimateMu.Lock()
			sugar.Info("MaxFeePerGas: ", gasEstimate.MaxFeePerGas, " GWEI")
			gasEstimateMu.Unlock()
		}
	}()
}

func CheckCycleWG(cycle uniswap.Cycle, pairToReserves *map[uniswap.Pair]uniswap.Reserve, executeChan chan uniswap.Cycle, pairToReservesMu *sync.Mutex, checkCounter *counter.TSCounter, gasEstimate *eth.EstimateFee, gasEstimateMu *sync.Mutex, balanceOf *big.Int, balanceOfMu *sync.Mutex, sugar *zap.SugaredLogger, wg *sync.WaitGroup) {
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
	wg.Done()
}
