package uniswap_router

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var _abi abi.ABI
var routerAddress common.Address

func SearchForAddLiquidity(data []byte, WETH common.Address) (common.Address, common.Address, big.Int, big.Int, bool) {
	var token0 common.Address
	var token1 common.Address
	var reserves0 big.Int
	var reserves1 big.Int

	if len(data) < 20 {
		return token0, token1, reserves0, reserves1, false
	}

	method, err := _abi.MethodById(data[:4])
	if err != nil {
		return token0, token1, reserves0, reserves1, false
	}

	m := make(map[string]interface{})
	method.Inputs.UnpackIntoMap(m, data[4:])

	if method.Name == "addLiquidity" {
		// fmt.Println(method.Name)
		// fmt.Println(m)
		token0 = m["tokenA"].(common.Address)
		token1 = m["tokenB"].(common.Address)
		reserves0 = *m["amountAMin"].(*big.Int)
		reserves1 = *m["amountBMin"].(*big.Int)
	} else if method.Name == "addLiquidityETH" {
		// fmt.Println(method.Name)
		// fmt.Println(m)
		token0 = WETH
		token1 = m["token"].(common.Address)
		reserves0 = *m["amountETHMin"].(*big.Int)
		reserves1 = *m["amountTokenMin"].(*big.Int)
	} else {
		return token0, token1, reserves0, reserves1, false
	}

	return token0, token1, reserves0, reserves1, true

}

func GetWETH(cl *ethclient.Client, ctx context.Context) (common.Address, error) {
	var wethAddress common.Address
	data, err := _abi.Pack("WETH")
	check(err)

	msg := ethereum.CallMsg{
		To:   &routerAddress,
		Data: data,
	}
	r, err := cl.CallContract(ctx, msg, nil)

	if err != nil {
		return wethAddress, err
	}

	m := make(map[string]interface{})
	err = _abi.UnpackIntoMap(m, "WETH", r)

	wethAddress = m[""].(common.Address)
	return wethAddress, nil

}

func GetAmountsOut(amountIn big.Int, path []common.Address, cl *ethclient.Client, ctx context.Context) []big.Int {
	// fmt.Println(path)
	data, err := _abi.Pack("getAmountsOut", &amountIn, &path)
	check(err)

	msg := ethereum.CallMsg{
		To:   &routerAddress,
		Data: data,
	}
	r, err := cl.CallContract(ctx, msg, nil)

	if err != nil {
		return make([]big.Int, 0)
	}

	m := make(map[string]interface{})
	err = _abi.UnpackIntoMap(m, "getAmountsOut", r)

	check(err)
	_amountOut := m["amounts"].([]*big.Int)
	amountOut := make([]big.Int, len(_amountOut))
	for k, _ := range _amountOut {
		amountOut[k] = *_amountOut[k]
	}
	return amountOut
}

func GetPath(data []byte) []common.Address {
	var path []common.Address
	if len(data) < 20 {
		return path
	}

	method, err := _abi.MethodById(data[:4])
	if err != nil {
		return path
	}

	m := make(map[string]interface{})
	method.Inputs.UnpackIntoMap(m, data[4:])

	switch m["path"].(type) {
	case []common.Address:
		path = m["path"].([]common.Address)
	}
	return path
}

func SwapExactETHForTokens(amountOutMin big.Int, path []common.Address, to common.Address, deadline uint) []byte {

	fmt.Println("AmmountOutMin:", amountOutMin.String())
	fmt.Println("Path:", path)
	fmt.Println("To:", to.String())
	fmt.Println("Deadline:", deadline)
	// os.Exit(0)

	d := new(big.Int)
	d.SetUint64(uint64(deadline))

	data, err := _abi.Pack("swapExactETHForTokens", &amountOutMin, &path, &to, &d)

	if err != nil {
		panic(err)
	}

	return data
}

func Init(_routerAddress string) {
	data, err := ioutil.ReadFile("./internal/uniswap_router/abi/router.json")
	check(err)
	abi, err := abi.JSON(bytes.NewReader(data))
	check(err)
	_abi = abi
	routerAddress = common.HexToAddress(_routerAddress)
}

func RemoveLiquidity(token common.Address, liquidity big.Int, amountTokenMin big.Int, amountETHMin big.Int, to common.Address, deadline big.Int) []byte {
	data, err := _abi.Pack("removeLiquidityETH", token, &liquidity, &amountTokenMin, &amountETHMin, to, &deadline)
	if err != nil {
		panic(err)
	}

	return data
}

func AddLiquidityETH(token common.Address, amountTokenDesired big.Int, amountTokenMin big.Int, amountETHMin big.Int, to common.Address, deadline big.Int) []byte {
	data, err := _abi.Pack("addLiquidityETH", token, &amountTokenDesired, &amountTokenMin, &amountETHMin, to, &deadline)
	if err != nil {
		panic(err)
	}

	return data
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
