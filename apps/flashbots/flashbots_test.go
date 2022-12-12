package flashbots

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/chenzhijie/go-web3"
	"github.com/chenzhijie/go-web3/types"
	"github.com/ethereum/go-ethereum/common"
	eTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/joho/godotenv"
)

const (
	mainnetInfuraProvider     = "https://mainnet.infura.io/v3/91ffab09868d430f9ce744c78d7ff427"
	goerliInfuraProvider      = "https://goerli.infura.io/v3/91ffab09868d430f9ce744c78d7ff427"
	goerliFlashbotMintNFTAddr = "0x20EE855E43A7af19E407E39E5110c2C1Ee41F64D"
)

func TestFlashbotSendBundleTx(t *testing.T) {

	godotenv.Load()
	signerKey := "ea0d86ce7b7c394ca92cafadb8c8b50e82820d79de32f993a78b16c0ab5b73ad"

	if len(signerKey) == 0 {
		t.Fatal("signer key or sender key is empty")
	}

	web3, err := web3.NewWeb3(goerliInfuraProvider)
	if err != nil {
		t.Fatal(err)
	}

	err = web3.Eth.SetAccount(signerKey)
	if err != nil {
		t.Fatal(err)
	}

	web3.Eth.SetChainId(5)

	currentBlockNumber, err := web3.Eth.GetBlockNumber()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("currentBlockNumber %v\n", currentBlockNumber)

	mintValue := web3.Utils.ToWei(0.03)

	mintNFTData, err := hex.DecodeString("1249c58b")
	if err != nil {
		t.Fatal(err)
	}

	bundleTxs := make([]*eTypes.Transaction, 0)

	gasLimit, err := web3.Eth.EstimateGas(&types.CallMsg{
		From:  web3.Eth.Address(),
		To:    common.HexToAddress(goerliFlashbotMintNFTAddr),
		Data:  mintNFTData,
		Value: types.NewCallMsgBigInt(mintValue),
	})
	if err != nil {
		t.Fatal(err)
	}
	nonce, err := web3.Eth.GetNonce(web3.Eth.Address(), nil)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("gaslimit %v nonce %v\n", gasLimit, nonce)

	mintNFTtx, err := web3.Eth.NewEIP1559Tx(
		common.HexToAddress(goerliFlashbotMintNFTAddr),
		mintValue, // 6a94d74f430000
		gasLimit,
		web3.Utils.ToGWei(0),  //
		web3.Utils.ToGWei(30), // b2d05e00
		mintNFTData,
		nonce,
	)
	if err != nil {
		t.Fatal(err)
	}

	bundleTxs = append(bundleTxs, mintNFTtx)

	fb, err := NewFlashBot(TestRelayURL, signerKey)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := fb.Simulate(
		bundleTxs,
		big.NewInt(int64(currentBlockNumber)),
		"latest",
	)

	if err != nil {
		t.Fatal(err)
	}
	egp, err := resp.EffectiveGasPrice()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Resp %s EffectiveGasPrice %v\n", resp, web3.Utils.FromGWei(egp))
	targetBlockNumber := big.NewInt(int64(currentBlockNumber) + 1)
	bundleResp, err := fb.SendBundle(
		bundleTxs,
		targetBlockNumber,
	)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("bundle resp %v\n", bundleResp)

	stat, err := fb.GetBundleStats(bundleResp.BundleHash, targetBlockNumber)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("bundle stat %v\n", stat)
}

func TestGetBundleStats(t *testing.T) {

	godotenv.Load()
	signerKey := "ea0d86ce7b7c394ca92cafadb8c8b50e82820d79de32f993a78b16c0ab5b73ad"

	if len(signerKey) == 0 {
		t.Fatal("signer key or sender key is empty")
	}

	fb, err := NewFlashBot(DefaultRelayURL, signerKey)
	if err != nil {
		t.Fatal(err)
	}

	bundleHash := "0x68610dccc7d19e8049b05c0f305c8f698616ad9090903b43536474147a08df53"
	targetBlockNumber := big.NewInt(int64(8119100))
	stat, err := fb.GetBundleStats(bundleHash, targetBlockNumber)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("bundle stat %v\n", stat)

}

func TestGetUserStats(t *testing.T) {

	godotenv.Load()
	signerKey := os.Getenv("signerKey")

	if len(signerKey) == 0 {
		t.Fatal("signer key or sender key is empty")
	}

	fb, err := NewFlashBot(TestRelayURL, signerKey)
	if err != nil {
		t.Fatal(err)
	}

	targetBlockNumber := big.NewInt(int64(6974433))
	fmt.Printf("%x\n", targetBlockNumber.Int64())
	stat, err := fb.GetUserStats(targetBlockNumber)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("user stat %v\n", stat)

}
