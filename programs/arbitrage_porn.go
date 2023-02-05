package programs

import (
	"arbitrage_go/blockchain"
	"arbitrage_go/config"
	"arbitrage_go/logging"
	"arbitrage_go/uniswap"
	"fmt"
	"math/big"
	"runtime"

	"github.com/ethereum/go-ethereum/common"
)

func cloneD(D map[common.Address]OptimalPath) map[common.Address]OptimalPath {
	clone := make(map[common.Address]OptimalPath)
	for k, o := range D {
		clone[k] = OptimalPath{
			path:     o.path,
			amountAt: o.amountAt,
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
	path     []common.Address
	amountAt *big.Int
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
		path:     []common.Address{WETH},
		amountAt: WETHIn,
	}

	for i := 0; i < MPL; i++ {
		oldD := cloneD(D)
		for x, optimalPathX := range oldD {
			for v, pairs := range E[x] {
				for _, pair := range pairs {
					amountAtV := uniswap.GetAmountOutToken(x, optimalPathX.amountAt, pair, pairToReserves[pair])

					_, exists := D[v]
					// no repeats and hop doesn't fail
					if !exists || new(big.Int).Sub(amountAtV, D[v].amountAt).Sign() == 1 {
						D[v] = OptimalPath{
							path:     append(optimalPathX.path, v),
							amountAt: amountAtV,
						}
					}
				}
			}
		}
	}
	fmt.Println("DONE")
	fmt.Println(D[WETH].path)
	fmt.Println(WETHIn, D[WETH].amountAt)
}
