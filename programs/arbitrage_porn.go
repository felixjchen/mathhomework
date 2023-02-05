package programs

import (
	"arbitrage_go/blockchain"
	"arbitrage_go/config"
	"arbitrage_go/logging"
	"arbitrage_go/uniswap"
	"fmt"
	"math/big"
	"runtime"

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

type OptimalPath struct {
	tokenPath []common.Address
	pairPath  []uniswap.Pair
	amountAt  *big.Int
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
	WETHIn := blockchain.GetWMETHBalance()
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
					// best and no repeats and hop doesn't fail
					best := new(big.Int).Sub(amountAtV, D[v].amountAt).Sign() == 1
					if best {
						D[v] = OptimalPath{
							tokenPath: append(optimalPathX.tokenPath, v),
							pairPath:  append(optimalPathX.pairPath, pair),
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