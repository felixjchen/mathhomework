package config

import "github.com/ethereum/go-ethereum/common"

var TOKEN_BLACKLIST = map[common.Address]bool{
	common.HexToAddress("0x1cd2528522A17B6Be63012fB63AE81f3e3e29D97"): true,
	common.HexToAddress("0xb072969BEfccBA697d87B2cB9CB00a1C78319E4D"): true,
	common.HexToAddress("0x72d731804Fb41d3e982d4B7E7e2c4642B3995651"): true,
	common.HexToAddress("0xd9575A55F0E4FB1123c8d95aeE1CE7b4432A647D"): true,
	// common.HexToAddress("0x663bD2f6BA806378640784DF2DBF109D775817CF"): true,
	// common.HexToAddress("0xb072969BEfccBA697d87B2cB9CB00a1C78319E4D"): true,
}
