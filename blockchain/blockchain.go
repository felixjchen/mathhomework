package blockchain

import (
	"arbitrage_go/config"
	"math/big"
	"sync"
	"time"

	"github.com/chenzhijie/go-web3"
	"github.com/chenzhijie/go-web3/eth"
)

var lock = &sync.Mutex{}

var Web3 *web3.Web3

func GetWeb3() *web3.Web3 {
	if Web3 == nil {
		lock.Lock()
		defer lock.Unlock()
		if Web3 == nil {
			newWeb3, err := web3.NewWeb3(config.Get().RPC_URL_HTTP)
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

func GetWMETHBalance() *big.Int {
	web3 := GetWeb3()
	wmatic, _ := web3.Eth.NewContract(config.WMATIC_ABI, config.Get().WETH_ADDRESS.Hex())
	balanceOfInterface, err := wmatic.Call("balanceOf", config.Get().BUNDLE_EXECUTOR_ADDRESS)
	for err != nil {
		time.Sleep(time.Second)
		balanceOfInterface, err = wmatic.Call("balanceOf", config.Get().BUNDLE_EXECUTOR_ADDRESS)
	}
	balanceOf, _ := balanceOfInterface.(*big.Int)
	return balanceOf
}

func GetGasEstimate() *eth.EstimateFee {
	newWeb3 := GetWeb3()
	gasEstimateMu := sync.Mutex{}
	gasEstimateMu.Lock()
	gasEstimate, err := newWeb3.Eth.EstimateFee()
	for err != nil {
		time.Sleep(time.Second)
		gasEstimate, err = newWeb3.Eth.EstimateFee()
	}
	return gasEstimate
}
