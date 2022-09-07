package uniswap

import (
	"arbitrage_go/blockchain"
	"arbitrage_go/config"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type Reserve struct {
	Reserve0           *big.Int
	Reserve1           *big.Int
	BlockTimestampLast *big.Int
}

// func TupleToReserve(tuple [3]*big.Int) Reserve {
// 	return Reserve{Reserve0: *tuple[0], Reserve1: *tuple[1], BlockTimestampLast: *tuple[2]}
// }

func UpdateReservesForPools(pools []Pool) map[Pool]Reserve {
	var wg sync.WaitGroup

	// mu protects reserves across threads
	var mu sync.Mutex
	// reserves := []Reserve{}
	poolToReserves := make(map[Pool]Reserve)

	web3 := blockchain.GetWeb3()

	poolAddresses := []common.Address{}
	for _, pool := range pools {
		poolAddresses = append(poolAddresses, pool.Address)
	}

	STEP_SIZE := 500
	for i := 0; i < len(pools); i += STEP_SIZE {
		j := i + STEP_SIZE
		if j > len(pools) {
			j = len(pools)
		}
		wg.Add(1)
		// Each thread gets a STEP_SIZE
		go func(start int, end int) {
			defer wg.Done()
			contract, err := web3.Eth.NewContract(config.UNISWAP_FLASH_QUERY_ABI, config.Get().FLASH_QUERY_ADDRESS.Hex())
			if err != nil {
				panic(err)
			}
			res, err := contract.Call("getReservesByPairs", poolAddresses[start:end])
			if err != nil {
				panic(err)
			}
			reserveTuples, ok := res.([][3]*big.Int)
			if !ok {
				fmt.Println("can not convert reserve")
			}
			// Create structs TODO MAP this
			reservesToAdd := []Reserve{}
			for _, arr := range reserveTuples {
				reservesToAdd = append(reservesToAdd, Reserve{Reserve0: arr[0], Reserve1: arr[1], BlockTimestampLast: arr[2]})
			}

			mu.Lock()
			for i, reserve := range reservesToAdd {
				poolToReserves[pools[start+i]] = reserve
			}
			mu.Unlock()
		}(i, j)
	}

	wg.Wait()
	return poolToReserves
}

func GetAmountOut(amountIn *big.Int, reserveIn *big.Int, reserveOut *big.Int) *big.Int {
	amountInWithFee := big.NewInt(0).Mul(amountIn, big.NewInt(997))
	numerator := big.NewInt(0).Mul(amountInWithFee, reserveOut)
	denominator := big.NewInt(0).Mul(reserveIn, big.NewInt(1000))
	denominator.Add(denominator, amountInWithFee)

	if denominator.Sign() == 0 {
		return big.NewInt(0)
	}

	return numerator.Div(numerator, denominator)
}

// TODO: Can optimise memory
// https://github.com/felixjchen/uniswap-arbitrage-analysis/blob/70366b389dca7eb0ba9a437598e2f72328a5613c/src/common.py#L229
func GetE0E1(R0 *big.Int, R1 *big.Int, R1_ *big.Int, R2 *big.Int) (*big.Int, *big.Int) {
	// 1000 * R1_ +  997 * R1
	A := big.NewInt(0).Mul(big.NewInt(1000), R1_)
	B := big.NewInt(0).Mul(big.NewInt(997), R1)
	denominator := big.NewInt(0).Add(A, B)

	if denominator.Sign() == 0 {
		return big.NewInt(0), big.NewInt(0)
	}

	var E0 *big.Int
	{
		// 1000 * R1_ * R0
		numerator := big.NewInt(0).Mul(R0, R1_)
		numerator.Mul(numerator, big.NewInt(1000))
		E0 = big.NewInt(0).Div(numerator, denominator)
	}

	var E1 *big.Int
	{
		// 997 * R1 * R2
		numerator := big.NewInt(0).Mul(R1, R2)
		numerator.Mul(numerator, big.NewInt(997))
		E1 = big.NewInt(0).Div(numerator, denominator)
	}

	return E0, E1
}

// TODO: Can optimise memory
// https://github.com/felixjchen/uniswap-arbitrage-analysis/blob/master/src/common.py#L171
func GetOptimalWethIn(E0 *big.Int, E1 *big.Int) *big.Int {
	// 1000 (sqrt(E0 * E1 * 997 / 1000) - E0)
	A := big.NewInt(0).Mul(E0, E1)
	A.Mul(A, big.NewInt(997))
	A.Mul(A, big.NewInt(1000))
	A.Sqrt(A)

	B := big.NewInt(0).Mul(big.NewInt(1000), E0)

	numerator := A.Sub(A, B)
	// 997
	denominator := big.NewInt(997)

	return big.NewInt(0).Div(numerator, denominator)
}
