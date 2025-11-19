package contracts

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	InkBridgeProxy = common.HexToAddress("0x2816cf15F6d2A220E789aA011D5EE4eB6c47FEbA")
	GateWayV3      = common.HexToAddress("0xDe090EfCD6ef4b86792e2D84E55a5fa8d49D25D2")
	AInkWlWETH     = common.HexToAddress("0x2B35eF056728BaFFaC103e3b81cB029788006EF9")
	WETH           = common.HexToAddress("0x4200000000000000000000000000000000000006")
)

type Delegate struct {
	client     *ethclient.Client
	bot        common.Address
	privateKey *ecdsa.PrivateKey
	safe       common.Address
	argus      common.Address
}

func NewDelegate(rpcUrl, delegatePrivateKey, safe, argus string) *Delegate {
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		log.Fatal(err)
	}
	privateKey, err := crypto.HexToECDSA(delegatePrivateKey)
	if err != nil {
		log.Fatal(err)
	}
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	return &Delegate{
		client:     client,
		bot:        address,
		privateKey: privateKey,
		safe:       common.HexToAddress(safe),
		argus:      common.HexToAddress(argus),
	}
}

func (d *Delegate) WithdrawETHFromGatewayV3(amount *big.Int) error {
	approvalData, err := buildAtokenApproval(GateWayV3, amount)
	if err != nil {
		return err
	}
	withdrawData, err := buildGatewayV3WithdrawETH(InkBridgeProxy, d.safe, amount)
	if err != nil {
		return err
	}

	var addrs []common.Address
	var values []*big.Int
	var datas [][]byte
	addrs = append(addrs, AInkWlWETH, GateWayV3)
	values = append(values, big.NewInt(0), big.NewInt(0))
	datas = append(datas, approvalData, withdrawData)

	safeExecData, err := buildSafeExecTransactions(addrs, values, datas)
	if err != nil {
		return err
	}

	txHash, err := d.SendTransaction(d.bot, d.argus, big.NewInt(0), safeExecData)
	if err != nil {
		return err
	}

	log.Printf("Transaction sent: %s", txHash.Hex())
	return nil
}

func (d *Delegate) SendTransaction(from, to common.Address, value *big.Int, data []byte) (common.Hash, error) {
	// Get nonce
	nonce, err := d.client.PendingNonceAt(context.Background(), d.bot)
	if err != nil {
		return common.Hash{}, err
	}

	// Get chain ID
	chainID, err := d.client.ChainID(context.Background())
	if err != nil {
		return common.Hash{}, err
	}

	// Estimate gas limit
	gasLimit, err := d.client.EstimateGas(context.Background(), ethereum.CallMsg{
		From:  from,
		To:    &to,
		Value: value,
		Data:  data,
	})
	if err != nil {
		return common.Hash{}, err
	}
	if gasLimit < uint64(3000000) {
		gasLimit = uint64(3000000)
	}
	fmt.Printf("gas limit: %d\n", gasLimit)

	// Get gas price suggestions
	gasTipCap, err := d.client.SuggestGasTipCap(context.Background())
	if err != nil {
		return common.Hash{}, err
	}

	// Get base fee from latest block
	header, err := d.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return common.Hash{}, err
	}
	baseFee := header.BaseFee

	// Set max fee per gas (base fee * 2 + tip)
	gasFeeCap := new(big.Int).Add(
		gasTipCap,
		new(big.Int).Mul(baseFee, big.NewInt(2)),
	)

	// Create EIP-1559 transaction
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &to,
		Value:     value,
		Data:      data,
	})

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewLondonSigner(chainID), d.privateKey)
	if err != nil {
		return common.Hash{}, err
	}

	// Send transaction
	err = d.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return common.Hash{}, err
	}

	return signedTx.Hash(), nil
}
