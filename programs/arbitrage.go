package programs

import (
	"arbitrage_go/blockchain"
	"arbitrage_go/counter"
	"arbitrage_go/database"
	"arbitrage_go/logging"
	"arbitrage_go/uniswap"
	"arbitrage_go/util"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

const MAX_CHECK_SIZE = 10000

// const MAX_QUERY_SIZE = 100000

// 0.002 ETHER
const BATCH_THRESHOLD = 2000000000000000

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

	RoughHopGasLimitMu := sync.Mutex{}
	RoughHopGasLimit := make(map[int]uniswap.RoughHopGasLimit)
	for i := 0; i < 10; i++ {
		RoughHopGasLimit[i] = uniswap.RoughHopGasLimit{0, time.Now()}
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
					go uniswap.CheckCycleWG(cycle, &pairToReserves, executeChan, &pairToReservesMu, checkCounter, gasEstimate, &gasEstimateMu, balanceOf, &balanceOfMu, &RoughHopGasLimitMu, &RoughHopGasLimit, sugar, &wg)
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
