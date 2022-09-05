package config

import "github.com/ethereum/go-ethereum/common"

var TOKEN_BLACKLIST = map[common.Address]bool{
	common.HexToAddress("0x1cd2528522A17B6Be63012fB63AE81f3e3e29D97"): true,
	common.HexToAddress("0xb072969BEfccBA697d87B2cB9CB00a1C78319E4D"): true,
}
