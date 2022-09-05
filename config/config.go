package config

import (
	"os"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
)

const PROD = true

type Config struct {
	init bool

	RPC_URL  string
	CHAIN_ID int64

	FLASH_QUERY_ADDRESS     common.Address
	BUNDLE_EXECUTOR_ADDRESS common.Address
	WETH_ADDRESS            common.Address

	UNISWAPV2_FACTORIES []common.Address

	PRIVATE_KEY string
}

var lock = &sync.Mutex{}

var GlobalConfig Config

func Get() *Config {

	// TODO ENV VARS
	godotenv.Load()

	if !GlobalConfig.init {
		lock.Lock()
		defer lock.Unlock()
		if !GlobalConfig.init {
			if PROD {
				RPC_URL := "https://polygon-mainnet.infura.io/v3/1de294ccc0da4f2ab105c9770ab3b962"
				CHAIN_ID := int64(137)

				FLASH_QUERY_ADDRESS := common.HexToAddress("0x9C7FfE06A4c5C58A5D60bC95baAb56F558A4dACf")     // Polygon
				BUNDLE_EXECUTOR_ADDRESS := common.HexToAddress("0x8Bdc9d868950E6993B2d0Aa8e56E1cEDa5140200") // Polygon

				WETH_ADDRESS := common.HexToAddress("0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270") //  Polygon

				QUICKSWAP_FACTORY_ADDRESS := common.HexToAddress("0x5757371414417b8C6CAad45bAeF941aBc7d3Ab32") // Mumbai and Polygon
				SUSHISWAP_FACTORY_ADDRESS := common.HexToAddress("0xc35DADB65012eC5796536bD9864eD8773aBc74C4") // Mumbai and Polygon
				DFYN_FACTORY_ADDRESS := common.HexToAddress("0xE7Fb3e833eFE5F9c441105EB65Ef8b261266423B")      // Polygon

				UNISWAPV2_FACTORIES := []common.Address{SUSHISWAP_FACTORY_ADDRESS, QUICKSWAP_FACTORY_ADDRESS, DFYN_FACTORY_ADDRESS}

				PRIVATE_KEY := os.Getenv("POLYGON_PRIVATE_KEY")

				GlobalConfig = Config{true, RPC_URL, CHAIN_ID, FLASH_QUERY_ADDRESS, BUNDLE_EXECUTOR_ADDRESS, WETH_ADDRESS, UNISWAPV2_FACTORIES, PRIVATE_KEY}
			} else {
				RPC_URL := "https://polygon-mumbai.infura.io/v3/1de294ccc0da4f2ab105c9770ab3b962"
				CHAIN_ID := int64(80001)

				FLASH_QUERY_ADDRESS := common.HexToAddress("0x8Bdc9d868950E6993B2d0Aa8e56E1cEDa5140200")     // Mumbai
				BUNDLE_EXECUTOR_ADDRESS := common.HexToAddress("0x444908F8f8f7fC7BAa948782A7e89785c61AeD7E") // Mumbai

				WETH_ADDRESS := common.HexToAddress("0x9c3C9283D3e44854697Cd22D3Faa240Cfb032889") //  Mumbai

				QUICKSWAP_FACTORY_ADDRESS := common.HexToAddress("0x5757371414417b8C6CAad45bAeF941aBc7d3Ab32") // Mumbai and Polygon
				SUSHISWAP_FACTORY_ADDRESS := common.HexToAddress("0xc35DADB65012eC5796536bD9864eD8773aBc74C4") // Mumbai and Polygon
				UNISWAPV2_FACTORIES := []common.Address{SUSHISWAP_FACTORY_ADDRESS, QUICKSWAP_FACTORY_ADDRESS}

				PRIVATE_KEY := os.Getenv("MUMBAI_PRIVATE_KEY")

				GlobalConfig = Config{true, RPC_URL, CHAIN_ID, FLASH_QUERY_ADDRESS, BUNDLE_EXECUTOR_ADDRESS, WETH_ADDRESS, UNISWAPV2_FACTORIES, PRIVATE_KEY}

			}
		}
	}

	return &GlobalConfig
}
