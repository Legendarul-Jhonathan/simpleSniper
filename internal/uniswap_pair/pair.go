package uniswap_pair

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var _abi abi.ABI

func DecodeEvent(vLog types.Log) {

	for _, topic := range vLog.Topics {
		event, err := _abi.EventByID(topic) //?????????
		if err != nil {
			// log.Fatalln(err)
			continue
		}

		if event == nil {
			// fmt.Println("No event id")
			continue
		}

		fmt.Println("---")
		fmt.Println("Event name:", event.Name)
		m := make(map[string]interface{})
		_abi.UnpackIntoMap(m, event.Name, vLog.Data)
		fmt.Println(m)
	}

}

func GetReserves(address common.Address, cl *ethclient.Client, ctx context.Context) (big.Int, big.Int, uint32) {
	data, err := _abi.Pack("getReserves")
	check(err)

	msg := ethereum.CallMsg{
		To:   &address,
		Data: data,
	}
	r, err := cl.CallContract(ctx, msg, nil)

	if err != nil {
		panic(err)
	}

	m := make(map[string]interface{})
	_abi.UnpackIntoMap(m, "getReserves", r)

	token0Reserve := m["reserve0"].(*big.Int)
	token1Reserve := m["reserve1"].(*big.Int)
	lastBlock := m["blockTimestampLast"].(uint32)

	return *token0Reserve, *token1Reserve, lastBlock

}

func GetAddressOfTokens(address common.Address, cl *ethclient.Client, ctx context.Context) (common.Address, common.Address, error) {

	var token0, token1 common.Address

	token0RequestData, err := _abi.Pack("token0")
	check(err)

	msg := ethereum.CallMsg{
		To:   &address,
		Data: token0RequestData,
	}

	token0ResponseData, err := cl.CallContract(ctx, msg, nil)
	check(err)

	token0Respone := make(map[string]interface{})

	err = _abi.UnpackIntoMap(token0Respone, "token0", token0ResponseData)
	if err != nil {
		return token0, token1, err
	}

	token0 = token0Respone[""].(common.Address)

	token1RequestData, err := _abi.Pack("token1")
	check(err)

	msg = ethereum.CallMsg{
		To:   &address,
		Data: token1RequestData,
	}

	token1ResponseData, err := cl.CallContract(ctx, msg, nil)
	check(err)

	token1Respone := make(map[string]interface{})

	err = _abi.UnpackIntoMap(token1Respone, "token1", token1ResponseData)
	if err != nil {
		return token0, token1, err
	}

	token1 = token1Respone[""].(common.Address)

	return token0, token1, nil
}

func BalanceOf(pairAddress common.Address, owner common.Address, cl *ethclient.Client, ctx context.Context) *big.Int {
	data, err := _abi.Pack("balanceOf", owner)
	check(err)

	msg := ethereum.CallMsg{
		To:   &pairAddress,
		Data: data,
	}
	r, err := cl.CallContract(ctx, msg, nil)

	m := make(map[string]interface{})
	_abi.UnpackIntoMap(m, "balanceOf", r)

	if err != nil {
		panic(err)
	}

	return m[""].(*big.Int)
}

func Approve(spender common.Address, value big.Int) []byte {
	data, err := _abi.Pack("approve", spender, &value)
	check(err)

	return data
}

func Init() {
	data, err := ioutil.ReadFile("./internal/uniswap_pair/abi/pair.json")
	check(err)
	abi, err := abi.JSON(bytes.NewReader(data))
	check(err)
	_abi = abi
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
