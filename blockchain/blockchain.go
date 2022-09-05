package blockchain

import (
	"arbitrage_go/config"
	"fmt"
	"sync"

	"github.com/chenzhijie/go-web3"
)

var lock = &sync.Mutex{}

var Web3 *web3.Web3

func GetWeb3() *web3.Web3 {
	if Web3 == nil {
		lock.Lock()
		defer lock.Unlock()
		if Web3 == nil {
			fmt.Println("Creating Web3 instance.")

			newWeb3, err := web3.NewWeb3(config.Get().RPC_URL)
			if err != nil {
				panic(err)
			}

			newWeb3.Eth.SetChainId(config.Get().CHAIN_ID)

			// TODO bye PK
			err = newWeb3.Eth.SetAccount("ea0d86ce7b7c394ca92cafadb8c8b50e82820d79de32f993a78b16c0ab5b73ad")
			if err != nil {
				panic(err)
			}

			Web3 = newWeb3
		}
	}

	return Web3
}
