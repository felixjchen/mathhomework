package programs

import (
	"arbitrage_go/config"
	"arbitrage_go/uniswap"
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

	allPools := uniswap.GetAllPools()
	allPoolsSet := uniswap.GetAllPoolsSet(allPools)
	// allPools = uniswap.FilterPools(tokenBlacklistFilter, allPools)
	sugar.Info("Got ", len(allPools), " pools")

	// pathing
	// adjacency list
	tokensToPools := make(map[common.Address][]uniswap.Pool)
	for _, pool := range allPools {
		tokensToPools[pool.Token0] = append(tokensToPools[pool.Token0], pool)
		tokensToPools[pool.Token1] = append(tokensToPools[pool.Token1], pool)
	}
	sugar.Info("Created Graph")

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
			_, exist := allPoolsSet[*txn.To()]
			if exist {

				txnData := txn.Data()
				sugar.Info(txn.Hash().Hex(), txnData)
				// set to generic pool
				pool, _ := web3Http.Eth.NewContract(config.PAIR_ABI, txn.To().Hex())
				method, _ := pool.Abi.MethodById(txnData)

				// https://gist.github.com/crazygit/9279a3b26461d7cb03e807a6362ec855
				inputsSigData := txnData[4:]
				inputsMap := make(map[string]interface{})
				if err := method.Inputs.UnpackIntoMap(inputsMap, inputsSigData); err != nil {
					sugar.Fatal(err)
				} else {
					fmt.Println(inputsMap)
				}

			}
		}
	}
}
