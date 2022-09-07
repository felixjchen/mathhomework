# arbitrage_go

https://docs.polygon.technology/docs/develop/network-details/network/

https://gist.github.com/harlow/dbcd639cf8d396a2ab73#file-worker_refactored-go

https://defiprime.com/polygon#decentralized-exchanges-on-polygon

https://medium.com/rungo/everything-you-need-to-know-about-packages-in-go-b8bac62b74cc

https://github.com/QuickSwap/quickswap-core/tree/master/contracts


## Go Web 3

https://github.com/chenzhijie/go-web3

## DEXES

<!-- https://defiprime.com/polygon#decentralized-exchanges-on-polygon -->

https://cryptoticker.io/en/top-5-dex-polygon/


## Uniswap

https://github.com/Uniswap/v2-periphery/blob/master/contracts/libraries/UniswapV2Library.sol

https://github.com/chenzhijie/go-web3/blob/master/examples/contract/erc20.go

# TODO 

volume optimisation + flashbots

threading

more dexes

remove uniswap in repo onchain

load abi from onchain folder

loading bars

unname functions


### OLD MULTICALL SNIPPED

```
// INEFFICIENT
// pool, err := web3.Eth.NewContract(config.PAIR_ABI, path[0].Address.String())
// if err != nil {
// 	panic(err)
// }

// firstData, err := pool.EncodeABI("swap", amount0Out, amount1Out, path[1].Address, []byte{})
// if err != nil {
// 	panic(err)
// }

// secondData, err := pool2.EncodeABI("swap", amount0Out, amount1Out, config.Get().BUNDLE_EXECUTOR_ADDRESS, []byte{})
// if err != nil {
// 	panic(err)
// }

// data, err := executor.EncodeABI("uniswapWeth", wethIn, big.NewInt(0), [2]common.Address{firstTarget, secondTarget}, [2][]byte{firstData, secondData})
// if err != nil {
// 	panic(err)
// }
```

### OLD PRICING


```
// wethIn := web3.Utils.ToWei(0.0001)

// price first hop
// wethReserve := poolToReserves[path[0]].Reserve0
// intermediateReserve := poolToReserves[path[0]].Reserve1
// if path[0].Token1 == weth {
// 	wethReserve = poolToReserves[path[0]].Reserve1
// 	intermediateReserve = poolToReserves[path[0]].Reserve0
// }
// intermediateAmount := uniswap.GetAmountOut(wethIn, wethReserve, intermediateReserve)

// // price second hop
// wethReserve = poolToReserves[path[1]].Reserve0
// intermediateReserve = poolToReserves[path[1]].Reserve1
// if path[1].Token1 == weth {
// 	wethReserve = poolToReserves[path[1]].Reserve1
// 	intermediateReserve = poolToReserves[path[1]].Reserve0
// }
// wethOut := uniswap.GetAmountOut(intermediateAmount, intermediateReserve, wethReserve)

// arbProfit := big.NewInt(0).Sub(wethOut, wethIn)
```