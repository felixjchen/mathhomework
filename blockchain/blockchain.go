package blockchain

import (
	"arbitrage_go/config"
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
			newWeb3, err := web3.NewWeb3(config.Get().RPC_URL)
			if err != nil {
				panic(err)
			}

			newWeb3.Eth.SetChainId(config.Get().CHAIN_ID)

			err = newWeb3.Eth.SetAccount(config.Get().PRIVATE_KEY)
			if err != nil {
				panic(err)
			}

			Web3 = newWeb3
		}
	}

	return Web3
}
