package programs

import (
	"arbitrage_go/config"
	"arbitrage_go/uniswap"
	"arbitrage_go/util"
	"fmt"

	"github.com/chenzhijie/go-web3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

func Mempool() {
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	sugar.Info("Started")

	// allPools := uniswap.GetAllPools()
	// // allPoolsSet := uniswap.GetAllPoolsSet(allPools)
	// // allPools = uniswap.FilterPools(tokenBlacklistFilter, allPools)
	// sugar.Info("Got ", len(allPools), " pools")

	// // pathing
	// // adjacency list
	// tokensToPools := make(map[common.Address][]uniswap.Pool)
	// for _, pool := range allPools {
	// 	tokensToPools[pool.Token0] = append(tokensToPools[pool.Token0], pool)
	// 	tokensToPools[pool.Token1] = append(tokensToPools[pool.Token1], pool)
	// }
	// sugar.Info("Created Graph")

	factoryPairMap := uniswap.GetFactoryPairMap()

	// Mempool watching
	incomingTxns := make(chan *types.Transaction)
	web3Http, _ := web3.NewWeb3(config.Get().RPC_URL_HTTP)
	web3WS, _ := web3.NewWeb3(config.Get().RPC_URL_WS)
	// https://github.com/chenzhijie/go-web3/blob/master/rpc/subscribe_test.go
	_, err := web3WS.Client.Subscribe("newPendingTransactions", func(data []byte) {
		// Get TXN
		var hash common.Hash
		hash.UnmarshalJSON(data)
		txn, _ := web3Http.Eth.GetTransactionByHash(hash)
		incomingTxns <- txn
	})
	if err != nil {
		sugar.Fatal(err)
	}

	for txn := range incomingTxns {
		// txn.to() is nil for contract creation
		if txn.To() != nil {
			// If this is a router call
			// TODO_LOW map here
			if util.Contains(config.Get().UNISWAPV2_ROUTER02S, *txn.To()) {
				txnData := txn.Data()
				sugar.Info(txn.Hash().Hex(), " ", txn.To())

				// TODO_MED set to generic router
				router, _ := web3Http.Eth.NewContract(config.UNISWAP_ROUTER_02_ABI, txn.To().Hex())
				method, _ := router.Abi.MethodById(txnData)

				// https://gist.github.com/crazygit/9279a3b26461d7cb03e807a6362ec855
				raw := make(map[string]interface{})
				if err := method.Inputs.UnpackIntoMap(raw, txnData[4:]); err != nil {
					sugar.Fatal(err)
				}

				if method.Name == "swapExactTokensForTokens" {
					args := uniswap.GetSwapExactTokensForTokensArgs(raw)
					factory := config.Get().ROUTER02_FACTORY_MAP[*txn.To()]
					pairMap := factoryPairMap[factory]

					amountsOut := uniswap.GetAmountsOut(pairMap, args.AmountIn, args.Path)

					fmt.Println(amountsOut)
				}
				// TODO_HIGH other funcs
			}
		}
	}
}
