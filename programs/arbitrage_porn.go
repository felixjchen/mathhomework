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
			diffAt:    o.diffAt,
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

func getSimulate(cycle uniswap.Cycle, amountIn *big.Int, pairToReserves *map[uniswap.Pair]uniswap.Reserve) bool {

	// NOOOO
	newWeb3, err := web3.NewWeb3(config.Get().RPC_URL_HTTP)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	newWeb3.Eth.SetChainId(config.Get().CHAIN_ID)
	newWeb3.Eth.SetAccount(config.Get().PRIVATE_KEY)

	amountsOut := uniswap.GetAmountsOutCycle(*pairToReserves, amountIn, cycle)
	targets := uniswap.GetCycleTargets(cycle)
	cycleAmountsOut := uniswap.GetCycleAmountsOut(cycle, amountsOut)
	executor, err := newWeb3.Eth.NewContract(config.BUNDLE_EXECTOR_ABI, config.Get().BUNDLE_EXECUTOR_ADDRESS.Hex())
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	data, err := executor.EncodeABI("hi_view", amountIn, targets, cycleAmountsOut)
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
	// fmt.Println(err)
	return err == nil
}

type OptimalPath struct {
	tokenPath []common.Address
	pairPath  []uniswap.Pair
	diffAt    *big.Int
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
	MPL := 3

	WETH := config.Get().WETH_ADDRESS
	balanceOf := blockchain.GetWETHBalance()
	pairToReserves := uniswap.GetReservesForPairs(allPairs)

	D := make(map[common.Address]OptimalPath)
	D[WETH] = OptimalPath{
		tokenPath: []common.Address{WETH},
		pairPath:  []uniswap.Pair{},
		diffAt:    new(big.Int),
	}

	for i := 1; i < MPL; i++ {
		fmt.Println("CHECKING MAX NUMBER OF HOPS: ", i)
		oldD := cloneD(D)
		for x, optimalPathX := range oldD {
			for v, pairs := range E[x] {
				for _, pair := range pairs {
					tokenPath := append(optimalPathX.tokenPath, v)
					pairPath := append(optimalPathX.pairPath, pair)
					cycle := uniswap.Cycle{
						Tokens: tokenPath,
						Edges:  pairPath,
					}
					E0, E1 := uniswap.GetE0E1ForCycle(cycle, pairToReserves)
					amountIn := uniswap.GetOptimalAmountIn(E0, E1)
					if big.NewInt(0).Sub(amountIn, balanceOf).Sign() == 1 {
						amountIn = balanceOf
					}
					if amountIn.Sign() != 1 {
						continue
					}

					amountsOut := uniswap.GetAmountsOutCycle(pairToReserves, amountIn, cycle)
					amountAtV := amountsOut[len(amountsOut)-1]
					diffAtV := new(big.Int).Sub(amountAtV, amountIn)
					if _, exists := D[v]; !exists {
						fmt.Println("FOUND NEW BEST PATH FOR ", v)
						D[v] = OptimalPath{
							tokenPath: append(optimalPathX.tokenPath, v),
							pairPath:  []uniswap.Pair{pair},
							diffAt:    diffAtV,
						}
						continue
					}
					// best and no repeats and simulate
					best := new(big.Int).Sub(diffAtV, D[v].diffAt).Sign() == 1
					noRepeats := true
					simulate := getSimulate(cycle, amountIn, &pairToReserves)
					if best && noRepeats && simulate {
						fmt.Println("FOUND NEW BEST PATH FOR ", v)
						D[v] = OptimalPath{
							tokenPath: tokenPath,
							pairPath:  pairPath,
							diffAt:    diffAtV,
						}
						continue
					}
				}
			}
		}
	}

	// Final hop
	fmt.Println("FINAL HOP")
	oldD := cloneD(D)
	for x, optimalPathX := range oldD {
		v := config.Get().WETH_ADDRESS
		for _, pair := range E[x][v] {
			tokenPath := append(optimalPathX.tokenPath, v)
			pairPath := append(optimalPathX.pairPath, pair)
			cycle := uniswap.Cycle{
				Tokens: tokenPath,
				Edges:  pairPath,
			}
			E0, E1 := uniswap.GetE0E1ForCycle(cycle, pairToReserves)
			amountIn := uniswap.GetOptimalAmountIn(E0, E1)
			if amountIn.Sign() != 1 {
				continue
			}
			if big.NewInt(0).Sub(amountIn, balanceOf).Sign() == 1 {
				sugar.Info("MIN ", amountIn, balanceOf)
				amountIn = balanceOf
			}
			// amountAtV := uniswap.GetAmountOutToken(x, amountIn, pair, pairToReserves[pair])
			amountsOut := uniswap.GetAmountsOutCycle(pairToReserves, amountIn, cycle)
			amountAtV := amountsOut[len(amountsOut)-1]
			diffAtV := new(big.Int).Sub(amountAtV, amountIn)
			if pair.Address == common.HexToAddress("0xA70a4EFC9902305FAB118F929CdF5c0d8d1f293A") || pair.Address == common.HexToAddress("0x1d0986d3496Ce1F457A0b03ACc556e9042879d01") {
				fmt.Println(cycle)
				fmt.Println(E0, E1)
				fmt.Println(amountIn)
				fmt.Println(amountAtV)
				fmt.Println(diffAtV)
			}
			// best and no repeats and simulate
			best := new(big.Int).Sub(diffAtV, D[v].diffAt).Sign() == 1
			noRepeats := true
			simulate := getSimulate(cycle, amountIn, &pairToReserves)
			// fmt.Println(best, utils.NewUtils().FromWei(diffAtV), utils.NewUtils().FromWei(D[v].diffAt))
			if best && noRepeats && simulate {
				fmt.Println("FOUND NEW BEST PATH FOR ", v)
				D[v] = OptimalPath{
					tokenPath: tokenPath,
					pairPath:  pairPath,
					diffAt:    diffAtV,
				}
				continue

			}
		}
	}
	optimalPathWETH := D[WETH]
	cycle := uniswap.Cycle{
		Tokens: optimalPathWETH.tokenPath,
		Edges:  optimalPathWETH.pairPath,
	}
	E0, E1 := uniswap.GetE0E1ForCycle(cycle, pairToReserves)
	amountIn := uniswap.GetOptimalAmountIn(E0, E1)
	if big.NewInt(0).Sub(amountIn, balanceOf).Sign() == 1 {
		sugar.Info("MIN ", amountIn, balanceOf)
		amountIn = balanceOf
	}

	fmt.Println("DONE")
	fmt.Println("TOKENPATH: ", D[WETH].tokenPath)
	fmt.Println("PAIRPATH: ", D[WETH].pairPath)
	fmt.Println("PROFIT ", utils.NewUtils().FromWei(new(big.Int).Sub(D[WETH].diffAt, amountIn)), "ETHER")
	// execute
	{
		web3 := blockchain.GetWeb3()

		// nonce, err := web3.Eth.GetNonce(web3.Eth.Address(), nil)
		// if err != nil {
		// 	sugar.Fatal(err)
		// }
		// nounceCounter := counter.NewTSCounter(nonce)

		// run bundle
		executor, err := web3.Eth.NewContract(config.BUNDLE_EXECTOR_ABI, config.Get().BUNDLE_EXECUTOR_ADDRESS.Hex())
		if err != nil {
			panic(err)
		}
		targets := uniswap.GetCycleTargets(cycle)
		amountsOut := uniswap.GetAmountsOutCycle(pairToReserves, amountIn, cycle)
		cycleAmountsOut := uniswap.GetCycleAmountsOut(cycle, amountsOut)
		data := uniswap.GetPayload(cycle, executor, amountIn, targets, cycleAmountsOut)

		call := &types.CallMsg{
			From: web3.Eth.Address(),
			To:   executor.Address(),
			Data: data,
			Gas:  types.NewCallMsgBigInt(big.NewInt(types.MAX_GAS_LIMIT)),
		}
		gasLimit, err := web3.Eth.EstimateGas(call)
		if err != nil {
			panic(err)
		}
		gasEstimate := blockchain.GetGasEstimate()
		gasTipCap := gasEstimate.MaxPriorityFeePerGas
		gasFeeCap := gasEstimate.MaxFeePerGas
		hash, err := web3.Eth.SendRawEIP1559Transaction(
			executor.Address(),
			new(big.Int),
			gasLimit,
			gasTipCap,
			gasFeeCap,
			data,
		)
		if err != nil {
			panic(err)
		} else {
			sugar.Info("tx hash: ", hash)
		}
	}
}
