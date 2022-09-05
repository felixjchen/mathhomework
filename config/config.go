package config

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

// godotenv.Load()

const PROD = false

type Config struct {
	init bool

	RPC_URL  string
	CHAIN_ID int64

	FLASH_QUERY_ADDRESS     common.Address
	BUNDLE_EXECUTOR_ADDRESS common.Address
	WETH_ADDRESS            common.Address

	UNISWAPV2_FACTORIES []common.Address

	// PRIVATE_KEY string
}

var lock = &sync.Mutex{}

var GlobalConfig Config

func Get() *Config {
	if !GlobalConfig.init {
		lock.Lock()
		defer lock.Unlock()
		if !GlobalConfig.init {
			fmt.Println("Creating single instance now.")

			if PROD {
				RPC_URL := "https://polygon-mainnet.infura.io/v3/1de294ccc0da4f2ab105c9770ab3b962"
				CHAIN_ID := int64(137)

				FLASH_QUERY_ADDRESS := common.HexToAddress("0x9C7FfE06A4c5C58A5D60bC95baAb56F558A4dACf")     // Polygon
				BUNDLE_EXECUTOR_ADDRESS := common.HexToAddress("0x8Bdc9d868950E6993B2d0Aa8e56E1cEDa5140200") // Polygon

				WETH_ADDRESS := common.HexToAddress("0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270") //  Polygon

				QUICKSWAP_FACTORY_ADDRESS := common.HexToAddress("0x5757371414417b8C6CAad45bAeF941aBc7d3Ab32") // Mumbai and Polygon
				SUSHISWAP_FACTORY_ADDRESS := common.HexToAddress("0xc35DADB65012eC5796536bD9864eD8773aBc74C4") // Mumbai and Polygon

				UNISWAPV2_FACTORIES := []common.Address{SUSHISWAP_FACTORY_ADDRESS, QUICKSWAP_FACTORY_ADDRESS}

				GlobalConfig = Config{true, RPC_URL, CHAIN_ID, FLASH_QUERY_ADDRESS, BUNDLE_EXECUTOR_ADDRESS, WETH_ADDRESS, UNISWAPV2_FACTORIES}
			} else {
				RPC_URL := "https://polygon-mumbai.infura.io/v3/1de294ccc0da4f2ab105c9770ab3b962"
				CHAIN_ID := int64(80001)

				FLASH_QUERY_ADDRESS := common.HexToAddress("0x8Bdc9d868950E6993B2d0Aa8e56E1cEDa5140200")     // Polygon
				BUNDLE_EXECUTOR_ADDRESS := common.HexToAddress("0x1f39F18d6b1e397D4E05A149148C9Fa9Bcc0eB67") // Mumbai

				WETH_ADDRESS := common.HexToAddress("0x9c3c9283d3e44854697cd22d3faa240cfb032889") //  Polygon

				QUICKSWAP_FACTORY_ADDRESS := common.HexToAddress("0x5757371414417b8C6CAad45bAeF941aBc7d3Ab32") // Mumbai and Polygon
				SUSHISWAP_FACTORY_ADDRESS := common.HexToAddress("0xc35DADB65012eC5796536bD9864eD8773aBc74C4") // Mumbai and Polygon
				UNISWAPV2_FACTORIES := []common.Address{SUSHISWAP_FACTORY_ADDRESS, QUICKSWAP_FACTORY_ADDRESS}

				GlobalConfig = Config{true, RPC_URL, CHAIN_ID, FLASH_QUERY_ADDRESS, BUNDLE_EXECUTOR_ADDRESS, WETH_ADDRESS, UNISWAPV2_FACTORIES}

			}
		}
	}

	return &GlobalConfig
}
