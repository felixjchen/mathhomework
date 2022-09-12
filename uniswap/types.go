package uniswap

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// TODO_MED cleanup types

type SwapExactTokensForTokensArgs struct {
	AmountIn     *big.Int
	AmountOutMin *big.Int
	Path         []common.Address
	To           common.Address
	Deadline     *big.Int
}

func GetSwapExactTokensForTokensArgs(args map[string]interface{}) SwapExactTokensForTokensArgs {
	amountIn := args["amountIn"].(*big.Int)
	amountOutMin := args["amountOutMin"].(*big.Int)
	path := args["path"].([]common.Address)
	to := args["to"].(common.Address)
	deadline := args["deadline"].(*big.Int)
	return SwapExactTokensForTokensArgs{amountIn, amountOutMin, path, to, deadline}
}
