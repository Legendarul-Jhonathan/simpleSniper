package custom_contractsV2

import (
	"bytes"
	"io/ioutil"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var _abi abi.ABI
var contractAddress common.Address

func ContractAddress() common.Address {
	return contractAddress
}

func ComposeBuyData(tokenAddress common.Address, minAmountOut *big.Int, feeForMiner *big.Int) []byte {

	data, err := _abi.Pack("buy", tokenAddress, minAmountOut, feeForMiner)
	check(err)

	return data
}

func Init() {
	data, err := ioutil.ReadFile("./internal/custom_contractsV2/abi/custom_contract2.json")
	check(err)
	abi, err := abi.JSON(bytes.NewReader(data))
	check(err)
	_abi = abi
	contractAddress = common.HexToAddress("0x6256195b125d61904bfc0df1c9af92d4293562b4") //MAIN-NET
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
