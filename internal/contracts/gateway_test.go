package contracts

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

var (
	safe                    = "0x5a372D431B99DB15ff6fdbf39Cf17dd8F0f6bDAb"
	bot                     = "0x725889fd807182fd29c55fb55557ef5b000d8bb2"
	l1Argus                 = "0x89ef9bad008133fa3cea3b8d6dba5612c0ea1f13"
	l2Argus                 = "0x6A7180F6217a1279646222d6B28Cc60C7FfCc995"
	privateKey              = ""
	l1StandardBridgeAddress = common.HexToAddress("0x88FF1e5b602916615391F55854588EFcBB7663f0")
	l2StandardBridgeAddress = common.HexToAddress("0x4200000000000000000000000000000000000010")
	localRPC                = "http://localhost:8545"
)

// 取ETH
func TestWithdrawETH(t *testing.T) {
	amount, _ := new(big.Int).SetString("100000000000000027464", 10)
	delegate := NewDelegate(localRPC, privateKey, safe, l2Argus)
	err := delegate.WithdrawETHFromGatewayV3(amount)
	assert.NoError(t, err)
}
