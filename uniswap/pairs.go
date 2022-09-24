package uniswap

import (
	"arbitrage_go/blockchain"
	"arbitrage_go/config"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

const STEP_SIZE int64 = 1000

// func GetAllPairsSet(pools []Pair) map[common.Address]bool {
// 	set := make(map[common.Address]bool)
// 	for _, pool := range pools {
// 		set[pool.Address] = true
// 	}
// 	return set
// }

func GetAllPairsForFactory(factory common.Address) []Pair {
	allPairs := []Pair{}
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}

	web3 := blockchain.GetWeb3()
	contract, err := web3.Eth.NewContract(config.UNISWAP_FLASH_QUERY_ABI, config.Get().FLASH_QUERY_ADDRESS.Hex())
	if err != nil {
		panic(err)
	}
	allPairsLengthInterface, err := contract.Call("getAllPairsLength", factory)
	for allPairsLengthInterface == nil {
		time.Sleep(1 * time.Second)
		fmt.Println("retrying getAllPairsLength, got", allPairsLengthInterface, factory, err)
		allPairsLengthInterface, _ = contract.Call("getAllPairsLength", factory)
	}
	allPairsLength, _ := allPairsLengthInterface.(*big.Int)

	for i := int64(0); i < allPairsLength.Int64(); i += STEP_SIZE {
		wg.Add(1)
		go func(i int64) {
			defer wg.Done()
			slice, _ := contract.Call("getPairsByIndexRange", factory, big.NewInt(i), big.NewInt(i+STEP_SIZE))
			// Just casted from interface{} to [][3] Address
			castedSlice, ok := slice.([][3]common.Address)
			for !ok {
				time.Sleep(2 * time.Second)
				fmt.Println("retrying getPairsByIndexRange, got", slice, i)
				slice, _ = contract.Call("getPairsByIndexRange", factory, big.NewInt(i), big.NewInt(i+STEP_SIZE))
				castedSlice, ok = slice.([][3]common.Address)
			}
			// TODO_LOW Create structs map this
			pairsToAdd := []Pair{}
			for _, arr := range castedSlice {
				pairsToAdd = append(pairsToAdd, Pair{Token0: arr[0], Token1: arr[1], Address: arr[2]})
			}
			mu.Lock()
			allPairs = append(allPairs, pairsToAdd...)
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	allPairs = FilterPairs(TokenBlacklistFilter, allPairs)
	return allPairs
}

// dexAddress => token0Address => token1Address => pairAddress
func GetFactoryPairMap() map[common.Address]map[common.Address]map[common.Address]Pair {
	factoryPairMap := make(map[common.Address]map[common.Address]map[common.Address]Pair)
	for _, factory := range config.Get().UNISWAPV2_FACTORIES {
		factoryPairMap[factory] = make(map[common.Address]map[common.Address]Pair)
	}
	for _, factory := range config.Get().UNISWAPV2_FACTORIES {
		pairs := GetAllPairsForFactory(factory)
		for _, pair := range pairs {
			factoryPairMap[factory][pair.Token0] = make(map[common.Address]Pair)
			factoryPairMap[factory][pair.Token1] = make(map[common.Address]Pair)
		}
		for _, pair := range pairs {
			// TODO_LOW suboptimal here
			factoryPairMap[factory][pair.Token0][pair.Token1] = pair
			factoryPairMap[factory][pair.Token1][pair.Token0] = pair
		}
	}
	return factoryPairMap
}

func GetAllPairsArray() []Pair {
	allPairsArray := []Pair{}
	for _, factory := range config.Get().UNISWAPV2_FACTORIES {
		pairsToAdd := GetAllPairsForFactory(factory)
		allPairsArray = append(allPairsArray, pairsToAdd...)
	}
	allPairsArray = FilterPairs(TokenBlacklistFilter, allPairsArray)
	return allPairsArray
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

func WethFilter(i Pair) bool {
	weth := config.Get().WETH_ADDRESS
	return i.Token0 == weth || i.Token1 == weth
}
func TokenBlacklistFilter(i Pair) bool {
	_, token0Blacklisted := config.TOKEN_BLACKLIST[i.Token0]
	_, token1Blacklisted := config.TOKEN_BLACKLIST[i.Token1]
	return !token0Blacklisted && !token1Blacklisted
}
