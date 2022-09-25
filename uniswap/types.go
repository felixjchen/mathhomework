package uniswap

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// TODO_MED cleanup types

type BatchCandidate struct {
	Cycle     Cycle
	NetProfit *big.Int
}

type SwapExactTokensForTokensArgs struct {
	AmountIn     *big.Int
	AmountOutMin *big.Int
	Path         []common.Address
	To           common.Address
	Deadline     *big.Int
}

type Pair struct {
	Token0  common.Address
	Token1  common.Address
	Address common.Address
}

type Reserve struct {
	Reserve0           *big.Int
	Reserve1           *big.Int
	BlockTimestampLast *big.Int
}

type Cycle struct {
	Tokens []common.Address
	Edges  []Pair
}

func GetSwapExactTokensForTokensArgs(args map[string]interface{}) SwapExactTokensForTokensArgs {
	amountIn := args["amountIn"].(*big.Int)
	amountOutMin := args["amountOutMin"].(*big.Int)
	path := args["path"].([]common.Address)
	to := args["to"].(common.Address)
	deadline := args["deadline"].(*big.Int)
	return SwapExactTokensForTokensArgs{amountIn, amountOutMin, path, to, deadline}
}
