package uniswap

import (
	"arbitrage_go/config"
	"fmt"
	"math/big"

	"github.com/chenzhijie/go-web3"
	"github.com/ethereum/go-ethereum/common"
)

const STEP_SIZE int64 = 200

type Pool struct {
	Token0  common.Address
	Token1  common.Address
	Address common.Address
}

func GetAllPools() []Pool {
	allPools := []Pool{}

	MUMBAI_URL := "https://polygon-mumbai.infura.io/v3/1de294ccc0da4f2ab105c9770ab3b962"
	web3, err := web3.NewWeb3(MUMBAI_URL)
	if err != nil {
		panic(err)
	}

	contract, err := web3.Eth.NewContract(config.UNISWAP_FLASH_QUERY_ABI, config.FLASH_QUERY_ADDRESS)
	if err != nil {
		panic(err)
	}
	var i int64 = 0
	for {
		slice, _ := contract.Call("getPairsByIndexRange", common.HexToAddress(config.QUICKSWAP_FACTORY_ADDRESS), big.NewInt(i), big.NewInt(i+STEP_SIZE))
		// Just casted from interface{} to [][3] Address
		castedSlice, ok := slice.([][3]common.Address)
		if !ok {
			fmt.Println("can not convert")
		}
		// Create structs
		poolsToAdd := []Pool{}
		for _, arr := range castedSlice {
			poolsToAdd = append(poolsToAdd, Pool{Token0: arr[0], Token1: arr[1], Address: arr[2]})
		}
		allPools = append(allPools, poolsToAdd...)
		i += STEP_SIZE
		if len(poolsToAdd) == 0 {
			break
		}
	}

	return allPools
}

func UpdateReservesForPools(pools []Pool) {
	MUMBAI_URL := "https://polygon-mumbai.infura.io/v3/1de294ccc0da4f2ab105c9770ab3b962"
	web3, err := web3.NewWeb3(MUMBAI_URL)
	if err != nil {
		panic(err)
	}

	contract, err := web3.Eth.NewContract(config.UNISWAP_FLASH_QUERY_ABI, config.FLASH_QUERY_ADDRESS)
	if err != nil {
		panic(err)
	}

	pairs := []common.Address{}
	for _, pool := range pools {
		pairs = append(pairs, pool.Address)
	}
	reserves, err := contract.Call("getReservesByPairs", pairs)
	if err != nil {
		panic(err)
	}

	fmt.Println(reserves)
}

func FilterPools(candidate func(Pool) bool, pools []Pool) []Pool {
	filteredPools := []Pool{}
	for _, pool := range pools {
		if candidate(pool) {
			filteredPools = append(filteredPools, pool)
		}
	}
	return filteredPools
}
