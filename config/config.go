package config

import (
	"os"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
)

// https://polygonscan.com/address/0x8Bdc9d868950E6993B2d0Aa8e56E1cEDa5140200#code 20cents here

// https://defillama.com/chain/Polygon

// GOD DEX
// MMF : https://mmfinance.gitbook.io/docs/smart-contracts/polygon-smart-contracts
// SUSHISWAP : https://docs.sushi.com/docs/Developers/Deployment%20Addresses
// APESWAP : https://apeswap.gitbook.io/apeswap-finance/where-dev/smart-contracts#polygon
// DFYN : https://docs.dfyn.network/technical/contracts
// JETSWAP : https://docs.jetswap.finance/contracts
// Gravity : https://inthenextversion.gitbook.io/gravity-finance/smart-contracts

// MEH ? Both have standard factories but weird routers
// Vulcan : https://dappradar.com/polygon/exchanges/vulcandex
// Elk Finance : https://docs.elk.finance/addresses/polygon

// BAD
// FIREBIRD_FACTORY_ADDRESS := common.HexToAddress("0x5De74546d3B86C8Df7FEEc30253865e1149818C8")  // Polygon WEIRD FEES
// Meshswap : https://docs.meshswap.fi/developers/contract

const PROD = true

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

	// TODO ENV VARS
	godotenv.Load()

	if !GlobalConfig.init {
		lock.Lock()
		defer lock.Unlock()
		if !GlobalConfig.init {
			if PROD {
				PRIVATE_KEY := os.Getenv("POLYGON_PRIVATE_KEY")

				RPC_URL_HTTP := "https://soft-still-moon.matic.discover.quiknode.pro/8e541ee43ee71b4ca15719a9527783f010f2892f/"
				RPC_URL_WS := "wss://ws-nd-769-892-956.p2pify.com/5d802cc6f6c7447316b9fa2684b79023"
				CHAIN_ID := int64(137)

				WETH_ADDRESS := common.HexToAddress("0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270") //  Polygon WMATIC
				// WETH_ADDRESS := common.HexToAddress("0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174") //  Polygon USDC

				FLASH_QUERY_ADDRESS := common.HexToAddress("0x9C7FfE06A4c5C58A5D60bC95baAb56F558A4dACf")     // Polygon
				BUNDLE_EXECUTOR_ADDRESS := common.HexToAddress("0x6962AD493D2fA1cA187075A2653d9CdCE4C21DD3") // Polygon

				QUICKSWAP_FACTORY_ADDRESS := common.HexToAddress("0x5757371414417b8C6CAad45bAeF941aBc7d3Ab32") // Mumbai and Polygon
				// MMF_FACTORY_ADDRESS := common.HexToAddress("0x7cFB780010e9C861e03bCbC7AC12E013137D47A5")       // Polygon
				SUSHISWAP_FACTORY_ADDRESS := common.HexToAddress("0xc35DADB65012eC5796536bD9864eD8773aBc74C4") // Mumbai and Polygon
				// APESWAP_FACTORY_ADDRESS := common.HexToAddress("0xCf083Be4164828f00cAE704EC15a36D711491284")   // Polygon
				DFYN_FACTORY_ADDRESS := common.HexToAddress("0xE7Fb3e833eFE5F9c441105EB65Ef8b261266423B") // Polygon
				// POLYCAT_FACTORY_ADDRESS := common.HexToAddress("0x477Ce834Ae6b7aB003cCe4BC4d8697763FF456FA")
				JETSWAP_FACTORY_ADDRESS := common.HexToAddress("0x668ad0ed2622C62E24f0d5ab6B6Ac1b9D2cD4AC7")
				GRAVITY_FACTORY_ADDRESS := common.HexToAddress("0x3ed75AfF4094d2Aaa38FaFCa64EF1C152ec1Cf20")
				// WEIRD
				VULCAN_FACTORY_ADDRESS := common.HexToAddress("0x293f45b6F9751316672da58AE87447d712AF85D7") // Polygon
				ELK_FACTORY_ADDRESS := common.HexToAddress("0xE3BD06c7ac7E1CeB17BdD2E5BA83E40D1515AF2a")
				UNISWAPV2_FACTORIES := []common.Address{QUICKSWAP_FACTORY_ADDRESS, SUSHISWAP_FACTORY_ADDRESS, DFYN_FACTORY_ADDRESS, JETSWAP_FACTORY_ADDRESS, GRAVITY_FACTORY_ADDRESS, VULCAN_FACTORY_ADDRESS, ELK_FACTORY_ADDRESS}

				QUICKSWAP_ROUTER02_ADDRESS := common.HexToAddress("0xa5E0829CaCEd8fFDD4De3c43696c57F7D7A678ff") // Mumbai and Polygon
				// MMF_ROUTER02_ADDRESS := common.HexToAddress("0x51aba405de2b25e5506dea32a6697f450ceb1a17")       // Polygon
				SUSHISWAP_ROUTER02_ADDRESS := common.HexToAddress("0x1b02dA8Cb0d097eB8D57A175b88c7D8b47997506") // Mumbai and Polygon
				// APESWAP_ROUTER02_ADDRESS := common.HexToAddress("0xC0788A3aD43d79aa53B09c2EaCc313A787d1d607")   // Polygon
				DFYN_ROUTER02_ADDRESS := common.HexToAddress("0xA102072A4C07F06EC3B4900FDC4C7B80b6c57429")
				// POLYCAT_ROUTER02_ADDRESS := common.HexToAddress("0x94930a328162957FF1dd48900aF67B5439336cBD")
				JETSWAP_ROUTER02_ADDRESS := common.HexToAddress("0x5C6EC38fb0e2609672BDf628B1fD605A523E5923") // Polygon
				GRAVITY_ROUTER02_ADDRESS := common.HexToAddress("0x57dE98135e8287F163c59cA4fF45f1341b680248") // Polygon
				UNISWAPV2_ROUTER02S := []common.Address{QUICKSWAP_ROUTER02_ADDRESS, SUSHISWAP_ROUTER02_ADDRESS, DFYN_ROUTER02_ADDRESS, JETSWAP_ROUTER02_ADDRESS}

				ROUTER02_FACTORY_MAP := make(map[common.Address]common.Address)
				ROUTER02_FACTORY_MAP[QUICKSWAP_ROUTER02_ADDRESS] = QUICKSWAP_FACTORY_ADDRESS
				ROUTER02_FACTORY_MAP[SUSHISWAP_ROUTER02_ADDRESS] = SUSHISWAP_FACTORY_ADDRESS
				ROUTER02_FACTORY_MAP[DFYN_ROUTER02_ADDRESS] = DFYN_FACTORY_ADDRESS
				// ROUTER02_FACTORY_MAP[POLYCAT_ROUTER02_ADDRESS] = POLYCAT_FACTORY_ADDRESS
				// ROUTER02_FACTORY_MAP[MMF_ROUTER02_ADDRESS] = MMF_FACTORY_ADDRESS
				// ROUTER02_FACTORY_MAP[APESWAP_ROUTER02_ADDRESS] = APESWAP_FACTORY_ADDRESS
				ROUTER02_FACTORY_MAP[JETSWAP_ROUTER02_ADDRESS] = JETSWAP_FACTORY_ADDRESS
				ROUTER02_FACTORY_MAP[GRAVITY_ROUTER02_ADDRESS] = GRAVITY_FACTORY_ADDRESS

				GlobalConfig = Config{PRIVATE_KEY, RPC_URL_HTTP, RPC_URL_WS, CHAIN_ID, WETH_ADDRESS, FLASH_QUERY_ADDRESS, BUNDLE_EXECUTOR_ADDRESS, UNISWAPV2_FACTORIES, UNISWAPV2_ROUTER02S, ROUTER02_FACTORY_MAP, true}
			} else {
				PRIVATE_KEY := os.Getenv("MUMBAI_PRIVATE_KEY")

				RPC_URL_HTTP := "https://polygon-mumbai.infura.io/v3/75719af97abe4d9aa033c21c65d33aaa"
				RPC_URL_WS := "wss://ws-nd-615-584-120.p2pify.com/9b5b74c9c7e211544e7b23eee465031d"
				CHAIN_ID := int64(80001)

				WETH_ADDRESS := common.HexToAddress("0x9c3C9283D3e44854697Cd22D3Faa240Cfb032889") //  Mumbai WMATIC
				// WETH_ADDRESS := common.HexToAddress("0xe6b8a5cf854791412c1f6efc7caf629f5df1c747") //  Mumbai USDC

				FLASH_QUERY_ADDRESS := common.HexToAddress("0x8ac54e383B37CdcB1176B1FE2f88bC385ecDDBeF")     // Mumbai
				BUNDLE_EXECUTOR_ADDRESS := common.HexToAddress("0xDdF1b141af0B740Eeaab45BCB9D29599ab756EA2") // Mumbai

				QUICKSWAP_FACTORY_ADDRESS := common.HexToAddress("0x5757371414417b8C6CAad45bAeF941aBc7d3Ab32") // Mumbai and Polygon
				SUSHISWAP_FACTORY_ADDRESS := common.HexToAddress("0xc35DADB65012eC5796536bD9864eD8773aBc74C4") // Mumbai and Polygon
				UNISWAPV2_FACTORIES := []common.Address{SUSHISWAP_FACTORY_ADDRESS, QUICKSWAP_FACTORY_ADDRESS}

				QUICKSWAP_ROUTER02_ADDRESS := common.HexToAddress("0xa5E0829CaCEd8fFDD4De3c43696c57F7D7A678ff") // Mumbai and Polygon
				SUSHISWAP_ROUTER02_ADDRESS := common.HexToAddress("0x1b02dA8Cb0d097eB8D57A175b88c7D8b47997506") // Mumbai and Polygon
				UNISWAPV2_ROUTER02S := []common.Address{QUICKSWAP_ROUTER02_ADDRESS, SUSHISWAP_ROUTER02_ADDRESS}

				ROUTER02_FACTORY_MAP := make(map[common.Address]common.Address)
				ROUTER02_FACTORY_MAP[QUICKSWAP_ROUTER02_ADDRESS] = QUICKSWAP_FACTORY_ADDRESS
				ROUTER02_FACTORY_MAP[SUSHISWAP_ROUTER02_ADDRESS] = SUSHISWAP_FACTORY_ADDRESS

				GlobalConfig = Config{PRIVATE_KEY, RPC_URL_HTTP, RPC_URL_WS, CHAIN_ID, WETH_ADDRESS, FLASH_QUERY_ADDRESS, BUNDLE_EXECUTOR_ADDRESS, UNISWAPV2_FACTORIES, UNISWAPV2_ROUTER02S, ROUTER02_FACTORY_MAP, true}

			}
		}
	}

	return &GlobalConfig
}
