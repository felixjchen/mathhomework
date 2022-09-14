package uniswap

import (
	"arbitrage_go/blockchain"
	"arbitrage_go/config"
	"arbitrage_go/util"
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

func CloneReserve(i Reserve) Reserve {
	Reserve0 := new(big.Int).Set(i.Reserve0)
	Reserve1 := new(big.Int).Set(i.Reserve1)
	BlockTimestampLast := new(big.Int).Set(i.BlockTimestampLast)
	return Reserve{Reserve0, Reserve1, BlockTimestampLast}
}

func ClonePairToReserves(pairToReserves map[Pair]Reserve) map[Pair]Reserve {
	clone := make(map[Pair]Reserve)
	for pair, reserve := range pairToReserves {
		clone[pair] = CloneReserve(reserve)
	}
	return clone
}

func GetReservesForPairs(pools []Pair) map[Pair]Reserve {
	var wg sync.WaitGroup

	// mu protects reserves across threads
	var mu sync.Mutex
	// reserves := []Reserve{}
	poolToReserves := make(map[Pair]Reserve)

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
	amountInWithFee := new(big.Int).Mul(amountIn, big.NewInt(997))
	numerator := new(big.Int).Mul(amountInWithFee, reserveOut)
	denominator := new(big.Int).Mul(reserveIn, big.NewInt(1000))
	denominator.Add(denominator, amountInWithFee)
	if denominator.Sign() == 0 {
		return new(big.Int)
	}
	return numerator.Div(numerator, denominator)
}

func GetAmountsOut(pairMap map[common.Address]map[common.Address]Pair, pairToReserves map[Pair]Reserve, amountIn *big.Int, path []common.Address) []*big.Int {
	amounts := []*big.Int{amountIn}
	for i := 0; i < len(path)-1; i++ {
		pair := pairMap[path[i]][path[i+1]]
		reserveIn, reserveOut := SortReserves(path[i], pair, pairToReserves[pair])
		amountOut := GetAmountOut(amounts[len(amounts)-1], reserveIn, reserveOut)
		amounts = append(amounts, amountOut)
	}
	return amounts
}

func GetAmountsOutCycle(pairToReserves map[Pair]Reserve, amountIn *big.Int, cycle Cycle) []*big.Int {
	amounts := []*big.Int{amountIn}
	for i := 0; i < len(cycle.Edges); i++ {
		pair := cycle.Edges[i]
		reserveIn, reserveOut := SortReserves(cycle.Tokens[i], pair, pairToReserves[pair])
		amountOut := GetAmountOut(amounts[len(amounts)-1], reserveIn, reserveOut)
		amounts = append(amounts, amountOut)
	}
	return amounts
}

// TODO_LOW: Can optimise memory
// https://github.com/felixjchen/uniswap-arbitrage-analysis/blob/70366b389dca7eb0ba9a437598e2f72328a5613c/src/common.py#L229
func GetE0E1(R0 *big.Int, R1 *big.Int, R1_ *big.Int, R2 *big.Int) (*big.Int, *big.Int) {
	// 1000 * R1_ +  997 * R1
	A := new(big.Int).Mul(big.NewInt(1000), R1_)
	B := new(big.Int).Mul(big.NewInt(997), R1)
	denominator := new(big.Int).Add(A, B)

	if denominator.Sign() == 0 {
		return new(big.Int), new(big.Int)
	}

	var E0 *big.Int
	{
		// 1000 * R1_ * R0
		numerator := new(big.Int).Mul(R0, R1_)
		numerator.Mul(numerator, big.NewInt(1000))
		E0 = new(big.Int).Div(numerator, denominator)
	}

	var E1 *big.Int
	{
		// 997 * R1 * R2
		numerator := new(big.Int).Mul(R1, R2)
		numerator.Mul(numerator, big.NewInt(997))
		E1 = new(big.Int).Div(numerator, denominator)
	}

	return E0, E1
}

func GetE0E1ForCycle(cycle Cycle, pairToReserves map[Pair]Reserve) (*big.Int, *big.Int) {
	E0, E1 := new(big.Int), new(big.Int)
	init := false
	// E edges, E-1 virtual pools
	for i := 0; i < len(cycle.Edges)-1; i++ {
		if init {
			R1_, R2 := SortReserves(cycle.Tokens[i+1], cycle.Edges[i+1], pairToReserves[cycle.Edges[i+1]])
			E0, E1 = GetE0E1(E0, E1, R1_, R2)
		} else {
			R0, R1 := SortReserves(cycle.Tokens[i], cycle.Edges[i], pairToReserves[cycle.Edges[i]])
			R1_, R2 := SortReserves(cycle.Tokens[i+1], cycle.Edges[i+1], pairToReserves[cycle.Edges[i+1]])
			E0, E1 = GetE0E1(R0, R1, R1_, R2)
			init = true
		}
	}
	return E0, E1
}

// TODO_LOW: Can optimise memory
// https://github.com/felixjchen/uniswap-arbitrage-analysis/blob/master/src/common.py#L171
func GetOptimalAmountIn(E0 *big.Int, E1 *big.Int) *big.Int {
	// 1000 (sqrt(E0 * E1 * 997 / 1000) - E0)
	A := new(big.Int).Mul(E0, E1)
	A.Mul(A, big.NewInt(997))
	A.Mul(A, big.NewInt(1000))
	A.Sqrt(A)

	B := new(big.Int).Mul(big.NewInt(1000), E0)

	numerator := A.Sub(A, B)
	// 997
	denominator := big.NewInt(997)

	return new(big.Int).Div(numerator, denominator)
}

func SortReserves(tokenIn common.Address, pair Pair, reserve Reserve) (*big.Int, *big.Int) {
	if pair.Token0 == tokenIn {
		return reserve.Reserve0, reserve.Reserve1
	} else {
		return reserve.Reserve1, reserve.Reserve0
	}
}

func GetCycleAmountsOut(cycle Cycle, amountsOut []*big.Int) ([]*big.Int, []*big.Int) {
	amounts0Out := []*big.Int{}
	amounts1Out := []*big.Int{}
	// A -> B -> C -> A
	//   b_   c_   a_
	for i, amountOut := range amountsOut[1:] {
		amount0Out := util.Ternary(cycle.Edges[i].Token0 == cycle.Tokens[i+1], amountOut, big.NewInt(0))
		amount1Out := util.Ternary(cycle.Edges[i].Token1 == cycle.Tokens[i+1], amountOut, big.NewInt(0))
		amounts0Out = append(amounts0Out, amount0Out)
		amounts1Out = append(amounts1Out, amount1Out)
	}
	return amounts0Out, amounts1Out
}

func GetCycleTargets(cycle Cycle) []common.Address {
	res := []common.Address{}
	for _, pair := range cycle.Edges {
		res = append(res, pair.Address)
	}
	return res
}
