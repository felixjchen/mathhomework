package uniswap

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// TODO_MED cleanup types

type SwapExactTokensForTokensArgs struct {
	amountIn     *big.Int
	amountOutMin *big.Int
	path         []common.Address
	to           common.Address
	deadline     *big.Int
}

func GetSwapExactTokensForTokensArgs(args map[string]interface{}) SwapExactTokensForTokensArgs {
	amountIn := args["amountIn"].(*big.Int)
	amountOutMin := args["amountOutMin"].(*big.Int)
	path := args["path"].([]common.Address)
	to := args["to"].(common.Address)
	deadline := args["deadline"].(*big.Int)
	return SwapExactTokensForTokensArgs{amountIn, amountOutMin, path, to, deadline}
}
