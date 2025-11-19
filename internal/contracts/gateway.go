package contracts

import (
	_ "embed"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

//go:embed abis/gateway_v3.abi.json
var gatewayV3ABI string

//go:embed abis/atoken.abi.json
var atokenABI string

func buildGatewayV3DepositETH(arg0, onBehalfOf common.Address, referralCode uint16) ([]byte, error) {
	gatewayV3, err := abi.JSON(strings.NewReader(gatewayV3ABI))
	if err != nil {
		return nil, err
	}
	data, err := gatewayV3.Pack("depositETH", arg0, onBehalfOf, referralCode)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func buildGatewayV3WithdrawETH(arg0, to common.Address, amount *big.Int) ([]byte, error) {
	gatewayV3, err := abi.JSON(strings.NewReader(gatewayV3ABI))
	if err != nil {
		return nil, err
	}
	data, err := gatewayV3.Pack("withdrawETH", arg0, amount, to)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func buildAtokenApproval(spender common.Address, amount *big.Int) ([]byte, error) {
	atoken, err := abi.JSON(strings.NewReader(atokenABI))
	if err != nil {
		return nil, err
	}
	data, err := atoken.Pack("approve", spender, amount)
	if err != nil {
		return nil, err
	}
	return data, nil
}
