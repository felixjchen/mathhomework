package config

import (
	"os"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
)

// https://polygonscan.com/address/0x8Bdc9d868950E6993B2d0Aa8e56E1cEDa5140200#code 20cents here

const PROD = false

type Config struct {
	init bool

	RPC_URL_HTTP string
	RPC_URL_WS   string

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
				RPC_URL_HTTP := "https://polygon-mainnet.infura.io/v3/1de294ccc0da4f2ab105c9770ab3b962"
				RPC_URL_WS := "https://polygon-mainnet.infura.io/v3/1de294ccc0da4f2ab105c9770ab3b962"

				CHAIN_ID := int64(137)

				FLASH_QUERY_ADDRESS := common.HexToAddress("0x9C7FfE06A4c5C58A5D60bC95baAb56F558A4dACf")     // Polygon
				BUNDLE_EXECUTOR_ADDRESS := common.HexToAddress("0x4040551f724F928E7bE00389dEa9C654163BFbb9") // Polygon

				WETH_ADDRESS := common.HexToAddress("0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270") //  Polygon

				QUICKSWAP_FACTORY_ADDRESS := common.HexToAddress("0x5757371414417b8C6CAad45bAeF941aBc7d3Ab32") // Mumbai and Polygon
				SUSHISWAP_FACTORY_ADDRESS := common.HexToAddress("0xc35DADB65012eC5796536bD9864eD8773aBc74C4") // Mumbai and Polygon
				DFYN_FACTORY_ADDRESS := common.HexToAddress("0xE7Fb3e833eFE5F9c441105EB65Ef8b261266423B")      // Polygon
				// FIREBIRD_FACTORY_ADDRESS := common.HexToAddress("0x5De74546d3B86C8Df7FEEc30253865e1149818C8")  // Polygon WEIRD FEES
				POLYCAT_FACTORY_ADDRESS := common.HexToAddress("0x477Ce834Ae6b7aB003cCe4BC4d8697763FF456FA")

				UNISWAPV2_FACTORIES := []common.Address{SUSHISWAP_FACTORY_ADDRESS, QUICKSWAP_FACTORY_ADDRESS, DFYN_FACTORY_ADDRESS, POLYCAT_FACTORY_ADDRESS}

				PRIVATE_KEY := os.Getenv("POLYGON_PRIVATE_KEY")

				GlobalConfig = Config{true, RPC_URL_HTTP, RPC_URL_WS, CHAIN_ID, FLASH_QUERY_ADDRESS, BUNDLE_EXECUTOR_ADDRESS, WETH_ADDRESS, UNISWAPV2_FACTORIES, PRIVATE_KEY}
			} else {
				RPC_URL_HTTP := "https://nd-615-584-120.p2pify.com/9b5b74c9c7e211544e7b23eee465031d"
				RPC_URL_WS := "wss://ws-nd-615-584-120.p2pify.com/9b5b74c9c7e211544e7b23eee465031d"
				CHAIN_ID := int64(80001)

				FLASH_QUERY_ADDRESS := common.HexToAddress("0x8Bdc9d868950E6993B2d0Aa8e56E1cEDa5140200")     // Mumbai
				BUNDLE_EXECUTOR_ADDRESS := common.HexToAddress("0xA742bD0f41A6bf845d7F46eAA5278358771Fc39e") // Mumbai
				// FOK 0xcbb45b429dfF1Bd9f975B17a878847f4d8bdf546
				WETH_ADDRESS := common.HexToAddress("0x9c3C9283D3e44854697Cd22D3Faa240Cfb032889") //  Mumbai

				QUICKSWAP_FACTORY_ADDRESS := common.HexToAddress("0x5757371414417b8C6CAad45bAeF941aBc7d3Ab32") // Mumbai and Polygon
				SUSHISWAP_FACTORY_ADDRESS := common.HexToAddress("0xc35DADB65012eC5796536bD9864eD8773aBc74C4") // Mumbai and Polygon
				UNISWAPV2_FACTORIES := []common.Address{SUSHISWAP_FACTORY_ADDRESS, QUICKSWAP_FACTORY_ADDRESS}

				PRIVATE_KEY := os.Getenv("MUMBAI_PRIVATE_KEY")

				GlobalConfig = Config{true, RPC_URL_HTTP, RPC_URL_WS, CHAIN_ID, FLASH_QUERY_ADDRESS, BUNDLE_EXECUTOR_ADDRESS, WETH_ADDRESS, UNISWAPV2_FACTORIES, PRIVATE_KEY}

			}
		}
	}

	return &GlobalConfig
}
