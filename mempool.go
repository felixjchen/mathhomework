package main

import (
	"arbitrage_go/config"
	"arbitrage_go/uniswap"
	"time"

	"github.com/chenzhijie/go-web3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

// func tokenBlacklistFilter(i uniswap.Pool) bool {
// 	_, token0Blacklisted := config.TOKEN_BLACKLIST[i.Token0]
// 	_, token1Blacklisted := config.TOKEN_BLACKLIST[i.Token1]
// 	return !token0Blacklisted && !token1Blacklisted
// }

func main() {
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
	web3Http, _ := web3.NewWeb3(config.Get().RPC_URL_HTTP)
	web3WS, _ := web3.NewWeb3(config.Get().RPC_URL_WS)
	incomingTxns := make(chan *types.Transaction)
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
				sugar.Info(txn.Hash().Hex(), txn.Data())
			}
		}
	}

	<-time.After(10 * time.Minute)

}
