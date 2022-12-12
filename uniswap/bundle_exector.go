package uniswap

import (
	"arbitrage_go/blockchain"
	"arbitrage_go/config"
	"math/big"

	"github.com/chenzhijie/go-web3/eth"
	"github.com/ethereum/go-ethereum/common"
)

func getCallPayload(cycle Cycle, executor *eth.Contract, amountIn *big.Int, ethAmountToCoinbase *big.Int, targets []common.Address, cycleAmountsOut [][2]*big.Int) []byte {
	web3 := blockchain.GetWeb3()
	payloads := [][]byte{}
	pool, err := web3.Eth.NewContract(config.PAIR_ABI, cycle.Edges[0].Address.String())
	if err != nil {
		panic(err)
	}
	for i := range cycle.Edges {
		amount0Out := cycleAmountsOut[i][0]
		amount1Out := cycleAmountsOut[i][1]
		if i == 0 {
			data, err := pool.EncodeABI("swap", amount0Out, amount1Out, cycle.Edges[1].Address, []byte{})
			if err != nil {
				panic(err)
			}
			payloads = append(payloads, data)
		} else if i == len(cycle.Edges)-1 {
			data, err := pool.EncodeABI("swap", amount0Out, amount1Out, config.Get().BUNDLE_EXECUTOR_ADDRESS, []byte{})
			if err != nil {
				panic(err)
			}
			payloads = append(payloads, data)
		} else {
			data, err := pool.EncodeABI("swap", amount0Out, amount1Out, cycle.Edges[i+1].Address, []byte{})
			if err != nil {
				panic(err)
			}
			payloads = append(payloads, data)
		}
	}
	data, err := executor.EncodeABI("hp", amountIn, ethAmountToCoinbase, targets, payloads)
	if err != nil {
		panic(err)
	}
	return data
}
func getInterfacePayload(executor *eth.Contract, amountIn *big.Int, targets []common.Address, cycleAmountsOut [][2]*big.Int) []byte {
	data, err := executor.EncodeABI("hi", amountIn, targets, cycleAmountsOut)
	if err != nil {
		panic(err)
	}
	return data
}

func GetPayload(cycle Cycle, executor *eth.Contract, amountIn *big.Int, ethAmountToCoinbase *big.Int, targets []common.Address, cycleAmountsOut [][2]*big.Int) []byte {
	if config.USE_PLAIN_PAYLOAD {
		return getCallPayload(cycle, executor, amountIn, ethAmountToCoinbase, targets, cycleAmountsOut)
	} else {
		return getInterfacePayload(executor, amountIn, targets, cycleAmountsOut)
	}
}
