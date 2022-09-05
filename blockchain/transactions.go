package blockchain

import (
	"math/big"

	"github.com/chenzhijie/go-web3/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	eTypes "github.com/ethereum/go-ethereum/core/types"
)

func SendRawEIP1559Transaction(
	e *eth.Eth,
	to common.Address,
	nonce uint64,
	amount *big.Int,
	gasLimit uint64,
	gasTipCap *big.Int,
	gasFeeCap *big.Int,
	data []byte,
) (common.Hash, error) {
	// nonce, err := e.GetNonce(e.Address(), nil)
	var hash common.Hash
	// if err != nil {
	// 	return hash, err
	// }
	dynamicFeeTx := &eTypes.DynamicFeeTx{
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &to,
		Value:     amount,
		Data:      data,
	}

	signedTx, err := eTypes.SignNewTx(e.GetPrivateKey(), eTypes.LatestSignerForChainID(e.GetChainId()), dynamicFeeTx)
	if err != nil {
		return hash, err
	}

	txData, err := signedTx.MarshalBinary()
	if err != nil {
		return hash, err
	}

	err = e.c.Call("eth_sendRawTransaction", &hash, hexutil.Encode(txData))

	return hash, err
}
