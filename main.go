package main

import (
	"fmt"
	"math/big"

	"arbitrage_go/constants"

	"github.com/chenzhijie/go-web3"
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

	fmt.Println("Contract address: ", contract.Address())
	res, err := contract.Call("getPairsByIndexRange", constants.QUICKSWAP_FACTORY_ADDRESS, big.NewInt(0), big.NewInt(4))
	if err != nil {
		panic(err)
	}

	fmt.Printf("res %v\n", res)
}
