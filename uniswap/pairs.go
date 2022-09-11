package uniswap

import (
	"arbitrage_go/blockchain"
	"arbitrage_go/config"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const STEP_SIZE int64 = 1000

type Pair struct {
	Token0  common.Address
	Token1  common.Address
	Address common.Address
}

func GetAllPairsSet(pools []Pair) map[common.Address]bool {
	set := make(map[common.Address]bool)
	for _, pool := range pools {
		set[pool.Address] = true
	}
	return set
}

func GetAllPairs() []Pair {
	allPairs := []Pair{}

	web3 := blockchain.GetWeb3()

	contract, err := web3.Eth.NewContract(config.UNISWAP_FLASH_QUERY_ABI, config.Get().FLASH_QUERY_ADDRESS.Hex())
	if err != nil {
		panic(err)
	}

	// Get all pools for all dexes
	for _, uniswappy_address := range config.Get().UNISWAPV2_FACTORIES {

		allPairsLengthInterface, _ := contract.Call("getAllPairsLength", uniswappy_address)
		allPairsLength, ok := allPairsLengthInterface.(*big.Int)
		if !ok {
			fmt.Println("can not convert allPairsLength")
		}

		for i := int64(0); i < allPairsLength.Int64(); i += STEP_SIZE {
			slice, _ := contract.Call("getPairsByIndexRange", uniswappy_address, big.NewInt(i), big.NewInt(i+STEP_SIZE))
			// Just casted from interface{} to [][3] Address
			castedSlice, ok := slice.([][3]common.Address)
			if !ok {
				fmt.Println("can not convert pool")
			}
			// Create structs TODO MAP this
			pairsToAdd := []Pair{}
			for _, arr := range castedSlice {
				pairsToAdd = append(pairsToAdd, Pair{Token0: arr[0], Token1: arr[1], Address: arr[2]})
			}
			allPairs = append(allPairs, pairsToAdd...)
		}
	}

	return allPairs
}

func FilterPairs(candidate func(Pair) bool, pools []Pair) []Pair {
	filteredPairs := []Pair{}
	for _, pool := range pools {
		if candidate(pool) {
			filteredPairs = append(filteredPairs, pool)
		}
	}
	return filteredPairs
}
