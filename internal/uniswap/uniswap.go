package uniswap

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var _abi abi.ABI
var uniswapAddress common.Address

func DecodeNewPoolEvent(vLog types.Log) []common.Address {
	var pairs []common.Address
	for _, topic := range vLog.Topics {
		event, err := _abi.EventByID(topic) //?????????
		if err != nil {
			// log.Fatalln(err)
			continue
		}

		if event == nil {
			fmt.Println("No event id")
			continue
		}

		fmt.Println("---")
		fmt.Println("Event name:", event.Name)

		m := make(map[string]interface{})
		_abi.UnpackIntoMap(m, event.Name, vLog.Data)
		fmt.Println(m)
		l := m["pair"].(common.Address)

		pairs = append(pairs, l)
	}

	return pairs

}

func GetPairsLenght(cl *ethclient.Client, ctx context.Context) big.Int {
	data, err := _abi.Pack("allPairsLength")
	check(err)

	msg := ethereum.CallMsg{
		To:   &uniswapAddress,
		Data: data,
	}
	r, err := cl.CallContract(ctx, msg, nil)

	check(err)

	m := make(map[string]interface{})
	err = _abi.UnpackIntoMap(m, "allPairsLength", r)

	check(err)

	return *m[""].(*big.Int)
}

func GetPairAddress(n int, cl *ethclient.Client, ctx context.Context) common.Address {
	valueIn := new(big.Int)
	valueIn, _ = valueIn.SetString(strconv.Itoa(n), 10)

	data, err := _abi.Pack("allPairs", valueIn)
	check(err)

	msg := ethereum.CallMsg{
		To:   &uniswapAddress,
		Data: data,
	}
	r, err := cl.CallContract(ctx, msg, nil)

	if err != nil {
		panic(err)
	}

	m := make(map[string]interface{})
	err = _abi.UnpackIntoMap(m, "allPairs", r)
	check(err)

	return m["pair"].(common.Address)
}

func FindPairAddress(token0 common.Address, token1 common.Address, cl *ethclient.Client, ctx context.Context) common.Address {

	data, err := _abi.Pack("getPair", token0, token1)
	check(err)

	msg := ethereum.CallMsg{
		To:   &uniswapAddress,
		Data: data,
	}
	r, err := cl.CallContract(ctx, msg, nil)

	if err != nil {
		panic(err)
	}

	m := make(map[string]interface{})
	err = _abi.UnpackIntoMap(m, "getPair", r)
	check(err)

	return m["pair"].(common.Address)
}

func CreatePair(token0 common.Address, token1 common.Address) []byte {
	data, err := _abi.Pack("createPair", token0, token1)
	check(err)

	return data
}

func Init(factoryAddress string) {
	data, err := ioutil.ReadFile("./internal/uniswap/abi/uniswap.json")
	check(err)
	abi, err := abi.JSON(bytes.NewReader(data))
	check(err)
	_abi = abi
	uniswapAddress = common.HexToAddress(factoryAddress)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
