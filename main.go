package main

import (
	"fmt"
	"math/big"

	"arbitrage_go/constants"

	"github.com/chenzhijie/go-web3"
	"github.com/ethereum/go-ethereum/common"
)

func main() {
	// change to your rpc provider
	MUMBAI_URL := "https://polygon-mumbai.infura.io/v3/1de294ccc0da4f2ab105c9770ab3b962"
	// POLYGON_URL := "https://polygon-rpc.com"
	web3, err := web3.NewWeb3(MUMBAI_URL)
	if err != nil {
		panic(err)
	}
	blockNumber, err := web3.Eth.GetBlockNumber()
	if err != nil {
		panic(err)
	}
	fmt.Println("Current block number: ", blockNumber)

	contract, err := web3.Eth.NewContract(constants.UNISWAP_FLASH_QUERY_ABI, constants.FLASH_QUERY_ADDRESS)
	if err != nil {
		panic(err)
	}
	fmt.Println("Contract address: ", contract.Address())

	type Pairs [][3]common.Address

	// Collect all pairs
	allPairs := Pairs{}
	var i int64 = 0
	var step int64 = 200
	for {
		slice, _ := contract.Call("getPairsByIndexRange", common.HexToAddress(constants.QUICKSWAP_FACTORY_ADDRESS), big.NewInt(i), big.NewInt(i+step))

		pairs, ok := slice.([][3]common.Address)
		if !ok {
			fmt.Println("can not convert")
		}
		allPairs = append(allPairs, pairs...)
		i += step
		if len(pairs) == 0 {
			break
		}
	}

	// Filter to WETH pairs
	WETH := common.HexToAddress(constants.WETH_ADDRESS)
	wethPairs := Pairs{}
	for _, s := range allPairs {
		if s[0] == WETH || s[1] == WETH {
			wethPairs = append(wethPairs, s)
		}
	}
	fmt.Printf("WETH_PAIRS %v\n", wethPairs)
}
