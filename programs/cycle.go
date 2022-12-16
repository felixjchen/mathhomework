package programs

import (
	"arbitrage_go/config"
	"arbitrage_go/database"
	"arbitrage_go/logging"
	"arbitrage_go/uniswap"

	"github.com/ethereum/go-ethereum/common"
)

func FindCycles() {
	sugar := logging.GetSugar("cycle")

	allPairs := uniswap.GetAllPairsArray()
	sugar.Info("Got ", len(allPairs), " pairs")

	// adjacency list
	// token -> [edges]
	graph := make(map[common.Address][]uniswap.Pair)
	for _, pair := range allPairs {
		graph[pair.Token0] = append(graph[pair.Token0], pair)
		graph[pair.Token1] = append(graph[pair.Token1], pair)
	}
	sugar.Info("Created Graph")
	checkChan := make(chan uniswap.Cycle)

	go func() {
		db := database.NewDBConn(sugar)
		for cycle := range checkChan {
			db.AddCycle(cycle)
		}
	}()

	sugar.Info("START CYCLES")
	uniswap.GetCyclesToChan(config.Get().WETH_ADDRESS, graph, 4, checkChan)
	sugar.Info("DONE CYCLES")
}
