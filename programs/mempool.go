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

	allPairs := uniswap.GetAllPairsArray()
	sugar.Info("Got ", len(allPairs), " pairs")

	pairToReserves := uniswap.GetReservesForPairs(allPairs)
	sugar.Info("Updated ", len(allPairs), " Reserves")

	// adjacency list
	tokensToPairs := make(map[common.Address][]uniswap.Pair)
	for _, pair := range allPairs {
		tokensToPairs[pair.Token0] = append(tokensToPairs[pair.Token0], pair)
		tokensToPairs[pair.Token1] = append(tokensToPairs[pair.Token1], pair)
	}
	sugar.Info("Created Graph")

	pathes := uniswap.GetTwoHops(tokensToPairs)
	sugar.Info("Found ", len(pathes), " 2-hops")

	// TODO_MED wasted computation on second pair (its actually okay to try both directions)
	pairToPathes := make(map[uniswap.Pair][][2]uniswap.Pair)
	for _, path := range pathes {
		pairToPathes[path[0]] = append(pairToPathes[path[0]], path)
		pairToPathes[path[1]] = append(pairToPathes[path[1]], path)
	}

	factoryPairMap := uniswap.GetFactoryPairMap()
	sugar.Info("Updated factoryPairMap")

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

					// TODO_HIGH Trade is 2 hop (walletA -> Pair -> walletB)
					if len(args.Path) == 2 {
						amountsOut := uniswap.GetAmountsOut(pairMap, pairToReserves, args.AmountIn, args.Path)
						fmt.Println(amountsOut)

						updatedPair := pairMap[args.Path[0]][args.Path[1]]
						updatedPathes := pairToPathes[updatedPair]

						// simulate future state
						reserve := pairToReserves[updatedPair]

						for _, path := range updatedPathes {
							fmt.Println(path)
						}
					}

				}
				// TODO_HIGH other funcs
			}
		}
	}
}
