package uniswap

import (
	"arbitrage_go/config"
	"fmt"
	"math/big"

	"github.com/chenzhijie/go-web3"
	"github.com/ethereum/go-ethereum/common"
)

const STEP_SIZE int64 = 1000

type Pool struct {
	Token0  common.Address
	Token1  common.Address
	Address common.Address
}

func GetAllPools() []Pool {
	allPools := []Pool{}

	// FIX THIS CONFIG TODO
	// MUMBAI_URL := "https://polygon-mumbai.infura.io/v3/1de294ccc0da4f2ab105c9770ab3b962"
	POLYGON_URL := "https://polygon-rpc.com"
	web3, err := web3.NewWeb3(POLYGON_URL)
	if err != nil {
		panic(err)
	}

	contract, err := web3.Eth.NewContract(config.UNISWAP_FLASH_QUERY_ABI, config.FLASH_QUERY_ADDRESS)
	if err != nil {
		panic(err)
	}

	// Get all pools for all dexes
	for _, uniswappy_address := range config.UNISWAPPY_FACTORY_ADDRESSES {
		var i int64 = 0
		for {
			slice, _ := contract.Call("getPairsByIndexRange", common.HexToAddress(uniswappy_address), big.NewInt(i), big.NewInt(i+STEP_SIZE))
			// Just casted from interface{} to [][3] Address
			castedSlice, ok := slice.([][3]common.Address)
			if !ok {
				fmt.Println("can not convert pool")
			}
			// Create structs TODO MAP this
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
	}

	return allPools
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

type Reserve struct {
	Reserve0           *big.Int
	Reserve1           *big.Int
	BlockTimestampLast *big.Int
}

// func TupleToReserve(tuple [3]*big.Int) Reserve {
// 	return Reserve{Reserve0: *tuple[0], Reserve1: *tuple[1], BlockTimestampLast: *tuple[2]}
// }

func UpdateReservesForPools(pools []Pool) []Reserve {

	// FIX THIS CONFIG TODO
	// MUMBAI_URL := "https://polygon-mumbai.infura.io/v3/1de294ccc0da4f2ab105c9770ab3b962"
	POLYGON_URL := "https://polygon-rpc.com"
	web3, err := web3.NewWeb3(POLYGON_URL)

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
	res, err := contract.Call("getReservesByPairs", pairs)
	if err != nil {
		panic(err)
	}
	reserveTuples, ok := res.([][3]*big.Int)
	if !ok {
		fmt.Println("can not convert reserve")
	}

	// Create structs TODO MAP this
	reserves := []Reserve{}
	for _, arr := range reserveTuples {
		reserves = append(reserves, Reserve{Reserve0: arr[0], Reserve1: arr[1], BlockTimestampLast: arr[2]})
	}

	return reserves
}

func GetAmountOut(amountIn *big.Int, reserveIn *big.Int, reserveOut *big.Int) *big.Int {
	amountInWithFee := big.NewInt(0).Mul(amountIn, big.NewInt(997))
	numerator := big.NewInt(0).Mul(amountInWithFee, reserveOut)
	denominator := big.NewInt(0).Mul(reserveIn, big.NewInt(1000))
	denominator.Add(denominator, amountInWithFee)

	if denominator == big.NewInt(0) {
		return big.NewInt(0)
	}

	return numerator.Div(numerator, denominator)
}
