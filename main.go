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

	contractAddr := "0x3060307bada5370d2dbf2f52640cc91975394de3" // contract address
	contract, err := web3.Eth.NewContract(constants.UniswapFlashQueryABI, contractAddr)
	fmt.Println("Contract address: ", contract.Address())

	res, err := contract.Call("getPairsByIndexRange", common.HexToAddress("0x5757371414417b8C6CAad45bAeF941aBc7d3Ab32"), big.NewInt(0), big.NewInt(10))
	if err != nil {
		panic(err)
	}

	fmt.Printf("%x\n", res)
	fmt.Printf("res %v\n", res)
}
