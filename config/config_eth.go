package config

import (
	"os"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
)

const PROD = false
const USE_PLAIN_PAYLOAD = false

const MAX_CYCLE_SIZE = 5

// 0.002 ETHER
const BATCH_THRESHOLD = 2000000000000000
const MAX_BATCH_SIZE = 4

type Config struct {
	PRIVATE_KEY string

	RPC_URL_HTTP string
	RPC_URL_WS   string
	CHAIN_ID     int64

	WETH_ADDRESS common.Address

	FLASH_QUERY_ADDRESS     common.Address
	BUNDLE_EXECUTOR_ADDRESS common.Address

	UNISWAPV2_FACTORIES  []common.Address
	UNISWAPV2_ROUTER02S  []common.Address
	ROUTER02_FACTORY_MAP map[common.Address]common.Address

	init bool
}

var lock = &sync.Mutex{}

var GlobalConfig Config

func Get() *Config {

	godotenv.Load()

	if !GlobalConfig.init {
		lock.Lock()
		defer lock.Unlock()
		if !GlobalConfig.init {
			if PROD {
				PRIVATE_KEY := os.Getenv("MAINNET_PRIVATE_KEY")

				RPC_URL_HTTP := "https://mainnet.infura.io/v3/da3a20d54a954e25b4d1af6cf3439175"
				RPC_URL_WS := "wss://mainnet.infura.io/ws/v3/da3a20d54a954e25b4d1af6cf3439175"
				CHAIN_ID := int64(1)

				WETH_ADDRESS := common.HexToAddress("0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2")

				FLASH_QUERY_ADDRESS := common.HexToAddress("0x9C7FfE06A4c5C58A5D60bC95baAb56F558A4dACf")
				BUNDLE_EXECUTOR_ADDRESS := common.HexToAddress("0x3060307BaDA5370D2DbF2f52640CC91975394de3")

				// https://docs.uniswap.org/contracts/v2/reference/smart-contracts/factory
				UNISWAP_FACTORY := common.HexToAddress("0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f")
				//https://docs.sushi.com/docs/Developers/Deployment%20Addresses
				SUSHISWAP_FACTORY := common.HexToAddress("0xC0AEe478e3658e2610c5F7A4A2E1777cE9e4f2Ac")
				UNISWAPV2_FACTORIES := []common.Address{UNISWAP_FACTORY, SUSHISWAP_FACTORY}

				// EMPTY FOR NOW
				UNISWAPV2_ROUTER02S := []common.Address{}
				ROUTER02_FACTORY_MAP := make(map[common.Address]common.Address)

				GlobalConfig = Config{PRIVATE_KEY, RPC_URL_HTTP, RPC_URL_WS, CHAIN_ID, WETH_ADDRESS, FLASH_QUERY_ADDRESS, BUNDLE_EXECUTOR_ADDRESS, UNISWAPV2_FACTORIES, UNISWAPV2_ROUTER02S, ROUTER02_FACTORY_MAP, true}
			} else {
				PRIVATE_KEY := os.Getenv("GOERLI_PRIVATE_KEY")

				RPC_URL_HTTP := "https://goerli.infura.io/v3/da3a20d54a954e25b4d1af6cf3439175"
				RPC_URL_WS := "wss://goerli.infura.io/ws/v3/da3a20d54a954e25b4d1af6cf3439175"
				CHAIN_ID := int64(5)

				WETH_ADDRESS := common.HexToAddress("0xb4fbf271143f4fbf7b91a5ded31805e42b2208d6") //  Goerli WETH

				FLASH_QUERY_ADDRESS := common.HexToAddress("0xA0FCbeB907fefD8C83DAA67aa98608CfAB931A5D")
				BUNDLE_EXECUTOR_ADDRESS := common.HexToAddress("0x67d042A1Cf17a67467EeAFD315ff4d341E5b9e18")

				// https://docs.uniswap.org/contracts/v2/reference/smart-contracts/factory
				UNISWAP_FACTORY := common.HexToAddress("0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f")
				//https://docs.sushi.com/docs/Developers/Deployment%20Addresses
				SUSHISWAP_FACTORY := common.HexToAddress("0xc35DADB65012eC5796536bD9864eD8773aBc74C4")
				UNISWAPV2_FACTORIES := []common.Address{UNISWAP_FACTORY, SUSHISWAP_FACTORY}

				// EMPTY FOR NOW
				UNISWAPV2_ROUTER02S := []common.Address{}
				ROUTER02_FACTORY_MAP := make(map[common.Address]common.Address)

				GlobalConfig = Config{PRIVATE_KEY, RPC_URL_HTTP, RPC_URL_WS, CHAIN_ID, WETH_ADDRESS, FLASH_QUERY_ADDRESS, BUNDLE_EXECUTOR_ADDRESS, UNISWAPV2_FACTORIES, UNISWAPV2_ROUTER02S, ROUTER02_FACTORY_MAP, true}

			}
		}
	}

	return &GlobalConfig
}
