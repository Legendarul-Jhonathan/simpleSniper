package token

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gocarina/gocsv"
)

var _abi abi.ABI

type Token struct {
	ID       int    `csv:"ID"`
	Address  string `csv:"Address"`
	Name     string `csv:"Name"`
	Symbol   string `csv:"Symbol"`
	Decimals int32  `csv:"Decimals"`
}

type tokenDB []Token

var _tokenDB tokenDB

var _tokenDBMapID map[int]Token
var _tokenDBMapAddress map[string]Token

var cl *ethclient.Client
var ctx context.Context

// Keccak-256
func check(e error) {
	if e != nil {
		panic(e)
	}
}

func Init(_cl *ethclient.Client, _ctx context.Context) {
	data, err := ioutil.ReadFile("./internal/token/abi/token.json")
	check(err)
	abi, err := abi.JSON(bytes.NewReader(data))
	check(err)
	_abi = abi

	cl = _cl
	ctx = _ctx

	_tokenDBMapAddress = make(map[string]Token)
	_tokenDBMapID = make(map[int]Token)
	loadTokenDB()
}

func FindById(tokenId int) (Token, bool) {

	r, found := _tokenDBMapID[tokenId]

	if found == true {
		return r, true
	}

	var t Token
	return t, false
}

func GetName(_address string) (string, error) {
	data, err := _abi.Pack("name")
	check(err)

	address := common.HexToAddress(_address)

	msg := ethereum.CallMsg{
		To:   &address,
		Data: data,
	}
	r, err := cl.CallContract(ctx, msg, nil)

	if err != nil {
		return "", err
	}

	m := make(map[string]interface{})
	_abi.UnpackIntoMap(m, "name", r)

	if m[""] == nil {
		return "", nil
	}

	s := m[""].(string)
	return s, nil

}

func getSymbol(_address string) (string, error) {
	data, err := _abi.Pack("symbol")

	if err != nil {
		return "", err
	}

	address := common.HexToAddress(_address)

	msg := ethereum.CallMsg{
		To:   &address,
		Data: data,
	}
	r, err := cl.CallContract(ctx, msg, nil)

	if err != nil {
		return "", err
	}

	m := make(map[string]interface{})
	_abi.UnpackIntoMap(m, "symbol", r)

	if m[""] == nil {
		return "", nil
	}

	s := m[""].(string)
	return s, nil

}

func getDecimals(_address string) (int32, error) {
	data, err := _abi.Pack("decimals")
	check(err)

	address := common.HexToAddress(_address)

	msg := ethereum.CallMsg{
		To:   &address,
		Data: data,
	}
	r, err := cl.CallContract(ctx, msg, nil)

	if err != nil {
		return 0, err
	}

	m := make(map[string]interface{})
	_abi.UnpackIntoMap(m, "decimals", r)

	if m[""] == nil {
		return 0, nil
	}

	s := m[""].(uint8)
	ns := int32(s)
	return ns, nil

}

func HasLocker(address common.Address) bool {
	data, err := _abi.Pack("locker")
	check(err)

	msg := ethereum.CallMsg{
		To:   &address,
		Data: data,
	}
	r, err := cl.CallContract(ctx, msg, nil)

	if err != nil {
		return false
	}

	m := make(map[string]interface{})
	_abi.UnpackIntoMap(m, "locker", r)

	if m[""] == nil {
		return false
	}

	return true
}

func Find(address string) (Token, bool) {

	var t Token
	r, found := _tokenDBMapAddress[address]

	if found == true {
		return r, true
	}

	return t, false
}

func FindOrAdd(address string) (Token, bool) {
	IsToken(address)
	var t Token

	r, found := _tokenDBMapAddress[address]

	if found == true {
		return r, true
	}

	return t, false
}

//IsToken retrunes if a address is a token
func IsToken(add string) bool {
	var _, found = Find(add)

	if found == true {
		return true
	}

	_, err := tryGetToken(add)
	if err != nil {
		return true
	} else {
		return false
	}

}

func tryGetToken(add string) (Token, error) {
	var t Token

	symbol, err := getSymbol(add)

	if err != nil {
		return t, errors.New("Cannot get symbol")
	}

	name, err := GetName(add)
	if err != nil {
		return t, errors.New("Cannot get name")
	}

	decimal, err := getDecimals(add)
	if err != nil {
		return t, errors.New("Cannot get decimals")
	}

	if decimal > 64 {
		return t, errors.New("Decimal over 64, probably not a token")
	}

	t = Token{
		Address:  add,
		Name:     name,
		Symbol:   symbol,
		Decimals: decimal,
	}

	_tokenDB.add(t)
	SaveTokenDB()
	return t, nil
}

func (db tokenDB) add(t Token) {
	t.ID = len(_tokenDB) + 1
	_tokenDB = append(_tokenDB, t)
	_tokenDBMapAddress[t.Address] = t
	_tokenDBMapID[t.ID] = t

}

func (t Token) Print() {
	fmt.Println("ID:", t.ID)
	fmt.Println("Address: ", t.Address)
	fmt.Println("Decimals:", t.Decimals)
	fmt.Println("Symbol:", t.Symbol)
	fmt.Println("Name:", t.Name)
}

func SaveTokenDB() {
	tokenDBFile, err := os.OpenFile("./data/tokenDB.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer tokenDBFile.Close()

	err = gocsv.MarshalFile(&_tokenDB, tokenDBFile) // Use this to save the CSV back to the file
	if err != nil {
		panic(err)
	}

}

func loadMaps() {
	for _, t := range _tokenDB {
		_tokenDBMapAddress[t.Address] = t
		_tokenDBMapID[t.ID] = t
	}
}

func loadTokenDB() {
	tokenDBFile, err := os.OpenFile("./data/tokenDB.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer tokenDBFile.Close()

	fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "Loading token DB")

	err = gocsv.UnmarshalFile(tokenDBFile, &_tokenDB)
	if err != nil {
		panic(err)
	}
	loadMaps()
	fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "Loaded "+strconv.Itoa(len(_tokenDB))+" tokens")
}
