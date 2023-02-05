package programs

import (
	"arbitrage_go/blockchain"
	"arbitrage_go/config"
	"arbitrage_go/logging"
	"arbitrage_go/uniswap"
	"fmt"
	"log"
	"math/big"
	"runtime"

	"github.com/chenzhijie/go-web3"
	"github.com/chenzhijie/go-web3/types"
	"github.com/chenzhijie/go-web3/utils"
	"github.com/ethereum/go-ethereum/common"
)

func cloneD(D map[common.Address]OptimalPath) map[common.Address]OptimalPath {
	clone := make(map[common.Address]OptimalPath)
	for k, o := range D {
		// We might have cloning problems at pairPath
		clone[k] = OptimalPath{
			tokenPath: o.tokenPath,
			pairPath:  o.pairPath,
			amountAt:  o.amountAt,
		}
	}
	return clone
}

func getV(allPairs []uniswap.Pair) []common.Address {
	V := []common.Address{}

	vFound := make(map[common.Address]bool)
	for _, pair := range allPairs {
		if _, exists := vFound[pair.Token0]; !exists {
			vFound[pair.Token0] = true
			V = append(V, pair.Token0)
		}
		if _, exists := vFound[pair.Token1]; !exists {
			vFound[pair.Token1] = true
			V = append(V, pair.Token1)
		}
	}

	return V
}

func getE(V []common.Address, allPairs []uniswap.Pair) map[common.Address]map[common.Address][]uniswap.Pair {
	E := make(map[common.Address]map[common.Address][]uniswap.Pair)
	for _, v := range V {
		E[v] = make(map[common.Address][]uniswap.Pair)
	}
	for _, pair := range allPairs {
		E[pair.Token0][pair.Token1] = append(E[pair.Token0][pair.Token1], pair)
		E[pair.Token1][pair.Token0] = append(E[pair.Token1][pair.Token0], pair)
	}

	return E
}

func getSimulate(tokenPath []common.Address, pairPath []uniswap.Pair, amountIn *big.Int, pairToReserves *map[uniswap.Pair]uniswap.Reserve) bool {

	// NOOOO
	newWeb3, err := web3.NewWeb3(config.Get().RPC_URL_HTTP)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	newWeb3.Eth.SetChainId(config.Get().CHAIN_ID)
	newWeb3.Eth.SetAccount(config.Get().PRIVATE_KEY)

	cycle := uniswap.Cycle{
		Tokens: tokenPath,
		Edges:  pairPath,
	}
	amountsOut := uniswap.GetAmountsOutCycle(*pairToReserves, amountIn, cycle)
	targets := uniswap.GetCycleTargets(cycle)
	cycleAmountsOut := uniswap.GetCycleAmountsOut(cycle, amountsOut)
	executor, err := newWeb3.Eth.NewContract(config.BUNDLE_EXECTOR_ABI, config.Get().BUNDLE_EXECUTOR_ADDRESS.Hex())
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	data := uniswap.GetPayload(cycle, executor, amountIn, targets, cycleAmountsOut)
	call := &types.CallMsg{
		From: newWeb3.Eth.Address(),
		To:   executor.Address(),
		Data: data,
		Gas:  types.NewCallMsgBigInt(big.NewInt(types.MAX_GAS_LIMIT)),
	}
	_, err = newWeb3.Eth.EstimateGas(call)
	fmt.Println(err)
	return err == nil
}

type OptimalPath struct {
	tokenPath []common.Address
	pairPath  []uniswap.Pair
	amountAt  *big.Int
}

func ArbitragePornMain() {

	runtime.GOMAXPROCS(runtime.NumCPU())
	sugar := logging.GetSugar("arb_porn")

	allPairs := uniswap.GetAllPairsArray()
	sugar.Info("Got ", len(allPairs), " pairs")

	V := getV(allPairs)
	// x -> y -> pairs
	E := getE(V, allPairs)

	// each hop costs one
	MPL := 2

	WETH := config.Get().WETH_ADDRESS
	WETHIn := blockchain.GetWETHBalance()
	pairToReserves := uniswap.GetReservesForPairs(allPairs)

	D := make(map[common.Address]OptimalPath)
	D[WETH] = OptimalPath{
		tokenPath: []common.Address{WETH},
		pairPath:  []uniswap.Pair{},
		amountAt:  WETHIn,
	}

	for i := 0; i < MPL; i++ {
		oldD := cloneD(D)
		for x, optimalPathX := range oldD {
			for v, pairs := range E[x] {
				for _, pair := range pairs {
					amountAtV := uniswap.GetAmountOutToken(x, optimalPathX.amountAt, pair, pairToReserves[pair])

					if _, exists := D[v]; !exists {
						D[v] = OptimalPath{
							tokenPath: append(optimalPathX.tokenPath, v),
							pairPath:  []uniswap.Pair{pair},
							amountAt:  amountAtV,
						}
						continue
					}
					// data := uniswap.GetPayload(cycle, executor, amountIn, targets, cycleAmountsOut)
					// best and no repeats and simulate
					tokenPath := append(optimalPathX.tokenPath, v)
					pairPath := append(optimalPathX.pairPath, pair)
					best := new(big.Int).Sub(amountAtV, D[v].amountAt).Sign() == 1
					noRepeats := true
					simulate := getSimulate(tokenPath, pairPath, WETHIn, &pairToReserves)
					if best && noRepeats && simulate {
						fmt.Println("FOUND NEW BEST PATH FOR ", v)
						D[v] = OptimalPath{
							tokenPath: tokenPath,
							pairPath:  pairPath,
							amountAt:  amountAtV,
						}
						continue
					}
				}
			}
		}
	}
	fmt.Println("DONE")
	fmt.Println("TOKENPATH: ", D[WETH].tokenPath)
	fmt.Println("PAIRPATH: ", D[WETH].pairPath)
	fmt.Println("PROFIT ", utils.NewUtils().FromWei(new(big.Int).Sub(D[WETH].amountAt, WETHIn)), "ETHER")
}
