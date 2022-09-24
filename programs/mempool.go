package programs

func Mempool() {
	// logger, _ := zap.NewProduction()
	// sugar := logger.Sugar()
	// sugar.Info("Started")

	// allPairs := uniswap.GetAllPairsArray()
	// sugar.Info("Got ", len(allPairs), " pairs")

	// pairToReserves := uniswap.GetReservesForPairs(allPairs)
	// sugar.Info("Updated ", len(allPairs), " Reserves")

	// // adjacency list
	// tokensToPairs := make(map[common.Address][]uniswap.Pair)
	// for _, pair := range allPairs {
	// 	tokensToPairs[pair.Token0] = append(tokensToPairs[pair.Token0], pair)
	// 	tokensToPairs[pair.Token1] = append(tokensToPairs[pair.Token1], pair)
	// }
	// sugar.Info("Created Graph")

	// pathes := uniswap.GetTwoHops(tokensToPairs)
	// sugar.Info("Found ", len(pathes), " 2-hops")

	// // TODO_MED wasted computation on second pair (its actually okay to try both directions)
	// pairToPathes := make(map[uniswap.Pair][][2]uniswap.Pair)
	// for _, path := range pathes {
	// 	pairToPathes[path[0]] = append(pairToPathes[path[0]], path)
	// 	pairToPathes[path[1]] = append(pairToPathes[path[1]], path)
	// }

	// factoryPairMap := uniswap.GetFactoryPairMap()
	// sugar.Info("Updated factoryPairMap")

	// // Mempool watching
	// incomingTxns := make(chan *types.Transaction)
	// web3Http, _ := web3.NewWeb3(config.Get().RPC_URL_HTTP)
	// web3WS, _ := web3.NewWeb3(config.Get().RPC_URL_WS)
	// // https://github.com/chenzhijie/go-web3/blob/master/rpc/subscribe_test.go
	// _, err := web3WS.Client.Subscribe("newPendingTransactions", func(data []byte) {
	// 	// Get TXN
	// 	var hash common.Hash
	// 	hash.UnmarshalJSON(data)
	// 	txn, _ := web3Http.Eth.GetTransactionByHash(hash)
	// 	incomingTxns <- txn
	// })
	// if err != nil {
	// 	sugar.Fatal(err)
	// }

	// for txn := range incomingTxns {
	// 	// txn.to() is nil for contract creation
	// 	if txn != nil && txn.To() != nil {
	// 		// If this is a router call
	// 		// TODO_LOW map here
	// 		if util.Contains(config.Get().UNISWAPV2_ROUTER02S, *txn.To()) {
	// 			txnData := txn.Data()

	// 			// TODO_MED set to generic router
	// 			router, _ := web3Http.Eth.NewContract(config.UNISWAP_ROUTER_02_ABI, txn.To().Hex())
	// 			method, _ := router.Abi.MethodById(txnData)

	// 			// https://gist.github.com/crazygit/9279a3b26461d7cb03e807a6362ec855
	// 			raw := make(map[string]interface{})
	// 			if err := method.Inputs.UnpackIntoMap(raw, txnData[4:]); err != nil {
	// 				sugar.Fatal(err)
	// 			}

	// 			if method.Name == "swapExactTokensForTokens" {
	// 				args := uniswap.GetSwapExactTokensForTokensArgs(raw)
	// 				factory := config.Get().ROUTER02_FACTORY_MAP[*txn.To()]
	// 				pairMap := factoryPairMap[factory]
	// 				amountsOut := uniswap.GetAmountsOut(pairMap, pairToReserves, args.AmountIn, args.Path)

	// 				// For each updated pair
	// 				for i := 0; i < len(args.Path)-1; i++ {

	// 					pair := pairMap[args.Path[i]][args.Path[i+1]]

	// 					// simulate future state
	// 					// TODO_LOW inefficient cloning here
	// 					futurePairToReserves := uniswap.ClonePairToReserves(pairToReserves)
	// 					futureReserve := futurePairToReserves[pair]
	// 					if pair.Token0 == args.Path[i] {
	// 						futureReserve.Reserve0.Add(futureReserve.Reserve0, amountsOut[i])
	// 						futureReserve.Reserve1.Sub(futureReserve.Reserve1, amountsOut[i+1])
	// 					} else {
	// 						futureReserve.Reserve1.Add(futureReserve.Reserve1, amountsOut[i])
	// 						futureReserve.Reserve0.Sub(futureReserve.Reserve0, amountsOut[i+1])
	// 					}

	// 					gasTipCap := new(big.Int).Set(txn.GasTipCap())
	// 					gasTipCap.Sub(gasTipCap, big.NewInt(1))
	// 					gasFeeCap := txn.GasFeeCap()

	// 					updatedPathes := pairToPathes[pair]
	// 					sugar.Info("Incoming: ", txn.Hash().Hex(), " to: ", txn.To().Hex())
	// 					if len(updatedPathes) > 0 {
	// 						for _, path := range updatedPathes {
	// 							// sugar.Info("Backrunning: ", txn.Hash().Hex())
	// 							// sugar.Info("Pair: ", pair.Address)
	// 							// sugar.Info("AmountsOut: ", amountsOut)
	// 							// sugar.Info("Previous: ", pairToReserves[pair])
	// 							// sugar.Info("Future: ", futurePairToReserves[pair])
	// 							Arbitrage(path, futurePairToReserves, gasTipCap, gasFeeCap)
	// 						}
	// 					}
	// 				}
	// 			}
	// 			// TODO_HIGH other funcs
	// 		}
	// 	}
	// }
}
