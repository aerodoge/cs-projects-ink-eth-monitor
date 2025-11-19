package contracts

import (
	_ "embed"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

//go:embed abis/safe.abi.json
var safeABI string

type SafeCallData struct {
	Flag  *big.Int       `json:"flag"`  // uint256
	To    common.Address `json:"to"`    // address
	Value *big.Int       `json:"value"` // uint256
	Data  []byte         `json:"data"`  // bytes
	Hint  []byte         `json:"hint"`  // bytes
	Extra []byte         `json:"extra"` // bytes
}

func buildSafeExecTransaction(to common.Address, value *big.Int, data []byte) ([]byte, error) {
	callData := SafeCallData{
		Flag:  big.NewInt(0),
		To:    to,
		Value: value,
		Data:  data,
		Hint:  []byte{},
		Extra: []byte{},
	}

	safe, err := abi.JSON(strings.NewReader(safeABI))
	if err != nil {
		return nil, err
	}

	packedData, err := safe.Pack("execTransaction", callData)
	if err != nil {
		return nil, err
	}

	return packedData, nil
}

func buildSafeExecTransactions(addrs []common.Address, values []*big.Int, datas [][]byte) ([]byte, error) {
	safe, err := abi.JSON(strings.NewReader(safeABI))
	if err != nil {
		return nil, err
	}
	var callDatas []SafeCallData
	for i := range addrs {
		callData := SafeCallData{
			Flag:  big.NewInt(0),
			To:    addrs[i],
			Value: values[i],
			Data:  datas[i],
			Hint:  []byte{},
			Extra: []byte{},
		}
		callDatas = append(callDatas, callData)
	}

	data, err := safe.Pack("execTransactions", callDatas)
	if err != nil {
		return nil, err
	}

	return data, nil
}
