package config

import (
	"os"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
)

const PROD = false
const USE_PLAIN_PAYLOAD = false

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

				RPC_URL_HTTP := "https://mainnet.infura.io/v3/1de294ccc0da4f2ab105c9770ab3b962"
				RPC_URL_WS := "wss://ws-nd-769-892-956.p2pify.com/5d802cc6f6c7447316b9fa2684b79023"
				CHAIN_ID := int64(1)

				WETH_ADDRESS := common.HexToAddress("0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270") //  Polygon WMATIC
				// WETH_ADDRESS := common.HexToAddress("0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174") //  Polygon USDC

				FLASH_QUERY_ADDRESS := common.HexToAddress("0x9C7FfE06A4c5C58A5D60bC95baAb56F558A4dACf")     // Polygon
				BUNDLE_EXECUTOR_ADDRESS := common.HexToAddress("0x759eD4a2A3455FBC526c236ACbA04025B2f92113") // Polygon

				UNISWAPV2_FACTORIES := []common.Address{}

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
				BUNDLE_EXECUTOR_ADDRESS := common.HexToAddress("0x1CCC5b22Bbbc11Ed5DEEe4EBd88BefF4eAc6eEbe")

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
