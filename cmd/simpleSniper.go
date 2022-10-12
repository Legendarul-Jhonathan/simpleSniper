package main

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"fmt"
	"github.com/Legendarul-Jhonathan/simpleSniper/internal/conver"
	"github.com/Legendarul-Jhonathan/simpleSniper/internal/custom_contractsV2"
	"github.com/Legendarul-Jhonathan/simpleSniper/internal/token"
	"github.com/Legendarul-Jhonathan/simpleSniper/internal/uniswap"
	"github.com/Legendarul-Jhonathan/simpleSniper/internal/uniswap_pair"
	"github.com/Legendarul-Jhonathan/simpleSniper/internal/uniswap_router"
	"github.com/Legendarul-Jhonathan/simpleSniper/internal/ws"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
)

var cl *ethclient.Client
var ctx context.Context
var uniswapPairs []uniswapPair
var WETH common.Address

var target common.Address
var ourContract common.Address
var threshold big.Float
var privateKeys []string
var maxX int64
var amountToBuy float64

var HTTP_ADDRESS string
var WS_ADDRESS string
var IPC_ADDRESS string

var UNISWAP_FACTORY_ADDRESS string
var UNISWAP_ROUTER_ADDRESS string

var jobs chan sendTransactionJob
var jobResults chan common.Hash

type uniswapPair struct {
	PairAddress     common.Address `csv:"PairAddress"`
	Token0Address   common.Address `csv:"Token0Address"`
	Token0          token.Token    `csv:"-"`
	Token1Address   common.Address `csv:"Token1Address"`
	Token1          token.Token    `csv:"-"`
	Token0Reserve   big.Int        `csv:"Token0Reserve"`
	Token1Reserve   big.Int        `csv:"Token1Reserve"`
	LastTransaction uint32         `csv:"LastTransaction"`
	LastBlockUpdate uint64
}

type sendTransactionJob struct {
	privateKey string
	data       []byte
	gasLimit   uint64
	gasPrice   *big.Int
}

func main() {

	parseFlags()

	// Populate global variables
	UNISWAP_FACTORY_ADDRESS = "0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f"
	UNISWAP_ROUTER_ADDRESS = "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D"
	HTTP_ADDRESS = "127.0.0.1:8545"
	WS_ADDRESS = "127.0.0.1:8546"
	IPC_ADDRESS = "/srv/ethereum/geth.ipc"
	fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "UNISWAP_FACTORY_ADDRESS:", UNISWAP_FACTORY_ADDRESS)
	fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "UNISWAP_ROUTER_ADDRESS:", UNISWAP_ROUTER_ADDRESS)
	fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "HTTP_ADDRESS:", HTTP_ADDRESS)
	fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "WS_ADDRESS:", WS_ADDRESS)
	fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "IPC_ADDRESS:", IPC_ADDRESS)

	// Connect to geth
	_cl, err := ethclient.Dial(IPC_ADDRESS)
	if err != nil {
		panic(err)
	}
	cl = _cl
	ctx = context.Background()
	ws.Init(WS_ADDRESS)

	//Initialize various classes
	token.Init(cl, ctx)
	custom_contractsV2.Init()
	uniswap.Init(UNISWAP_FACTORY_ADDRESS)
	uniswap_pair.Init()
	uniswap_router.Init(UNISWAP_ROUTER_ADDRESS)

	//Create workers to send transactions from
	fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "Starting", len(privateKeys), "workers")
	jobs = make(chan sendTransactionJob, len(privateKeys))
	jobResults = make(chan common.Hash, len(privateKeys))
	for w := 1; w <= len(privateKeys); w++ {
		go sendTransactionWorker(jobs, jobResults)
		fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "Created worker", w)
	}

	// Get current block
	startBlock, err := cl.BlockNumber(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "Current block:", startBlock)

	//Fetching the WETH address
	WETH = getWETHAddress()
	fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "WETH address is", WETH.String())

	//Wait for new transaction
	NewTransactionChan := make(chan ws.WsResponse)
	go ws.NewPoolTransaction(NewTransactionChan)
	go waitForNewTrx(NewTransactionChan)

	dontExitProgramm()

}

func dontExitProgramm() {
	for {
		time.Sleep(10 * time.Millisecond)
	}
}

func waitForNewTrx(newHeadChan chan ws.WsResponse) {
	for {
		select {
		case data := <-newHeadChan:
			if data.Params.Result != "" {
				txHash := common.HexToHash(data.Params.Result)
				tx, _, _ := cl.TransactionByHash(ctx, txHash)
				searchTrx(tx)
			}

		default:
			t, _ := time.ParseDuration("10ms")
			time.Sleep(t)
		}
	}
}

func searchTrx(trx *types.Transaction) {

	//Trsansaction to uniswap router
	if trx != nil && trx.To() != nil && trx.To().String() == UNISWAP_ROUTER_ADDRESS {

		//Search for addLiquidity
		token0, token1, reserves0, reserves1, found := uniswap_router.SearchForAddLiquidity(trx.Data(), WETH)

		if found == true {

			var wethAddedLiquidity big.Int
			var otherTokenAddress common.Address
			var otherTokenAddedLiquidity big.Int

			fmt.Println("")
			fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "Add liquidity tx hash:", trx.Hash())

			if token0 == WETH {
				wethAddedLiquidity = reserves0
				otherTokenAddress = token1
				otherTokenAddedLiquidity = reserves1
			} else if token1 == WETH {
				wethAddedLiquidity = reserves1
				otherTokenAddress = token0
				otherTokenAddedLiquidity = reserves0
			} else {
				fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "Pool does not have WETH, skipping...")
				return
			}

			wethA := convert.SetDecimal(wethAddedLiquidity, 18)

			fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "WETH liquidity added:", wethA.String())

			if wethA.Cmp(&threshold) >= 0 && otherTokenAddress == target {

				expectedLiquidityWETH := new(big.Int)
				expectedLiquidityOtherToken := new(big.Int)

				expectedLiquidityWETH = &wethAddedLiquidity
				expectedLiquidityOtherToken = &otherTokenAddedLiquidity

				fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "expectedLiquidityWETH:", expectedLiquidityWETH)
				fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "expectedLiquidityOtherToken:", expectedLiquidityOtherToken)

				amountIn := getBalance(ourContract)
				amountOut := getAmountsOut(amountIn, expectedLiquidityWETH, expectedLiquidityOtherToken)

				maxXBigInt := new(big.Int)
				maxXBigInt.SetInt64(maxX)
				amountOutMin := new(big.Int)
				amountOutMin = amountOutMin.Div(amountOut, maxXBigInt)

				fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "amountIn:", amountIn.String())
				fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "amountOut:", amountOut.String())
				fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "amountOutMin", amountOutMin.String())
				fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "maxX", maxXBigInt.String())

				var minerFee big.Int
				minerFee.SetInt64(0)
				cData := custom_contractsV2.ComposeBuyData(otherTokenAddress, amountOutMin, &minerFee)

				gasLimit := uint64(500000) // in units

				for j := 1; j <= len(privateKeys); j++ {
					var job sendTransactionJob
					job.data = cData
					job.gasLimit = gasLimit
					job.gasPrice = trx.GasPrice()
					job.privateKey = privateKeys[j-1]
					jobs <- job
				}
				close(jobs)

				for a := 1; a <= len(privateKeys); a++ {
					h := <-jobResults
					fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "Tx hash from worker:", h)
				}

				fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "Transactions sent")
				fmt.Println("My work here is done, good luck")
				fmt.Println("Exiting...")
				os.Exit(0)
			}

		}

	}

}

func sendTransactionWorker(jobs <-chan sendTransactionJob, results chan<- common.Hash) {
	for j := range jobs {
		hash := sendTransactionFromWorker(j.data, j.gasLimit, j.gasPrice, j.privateKey)
		results <- hash
	}
}

func getBalance(target common.Address) *big.Int {
	balance, err := cl.BalanceAt(ctx, target, nil)
	if err != nil {
		panic(err)
	}

	return balance
}

func getAmountsOut(amountIn *big.Int, reserveIn *big.Int, reserveOut *big.Int) *big.Int {
	fee := new(big.Int)
	fee.SetInt64(997)
	oneK := new(big.Int)
	oneK.SetInt64(1000)

	amountInWithFee := new(big.Int)
	amountInWithFee = amountInWithFee.Mul(amountIn, fee)
	numerator := new(big.Int)
	numerator = numerator.Mul(amountInWithFee, reserveOut)
	denominator := reserveIn.Mul(reserveIn, oneK)
	denominator = denominator.Add(denominator, amountInWithFee)
	amountOut := numerator.Div(numerator, denominator)

	return amountOut
}

func sendTransactionFromWorker(data []byte, gasLimit uint64, gasPrice *big.Int, pKey string) common.Hash {
	privateKey, err := crypto.HexToECDSA(pKey)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := cl.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}

	toAddress := ourContract

	ethValue := new(big.Int)
	ethValue.SetInt64(0)

	tx := types.NewTransaction(nonce, toAddress, ethValue, gasLimit, gasPrice, data)

	chainID, err := cl.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	ts := types.Transactions{signedTx}
	rawTxBytes := ts.GetRlp(0)

	rlp.DecodeBytes(rawTxBytes, &tx)
	timeSendTransaction1 := time.Now()

	err = cl.SendTransaction(context.Background(), tx)

	if err != nil {
		fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "!!! Tx error:", err)
		return common.Hash{}
	}

	timeSendTransaction2 := time.Now()
	diff := timeSendTransaction2.Sub(timeSendTransaction1)
	fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "SendTransaction duration:", diff)
	fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "Tx hash:", tx.Hash().Hex())
	return tx.Hash()
}

func getWETHAddress() common.Address {
	address, err := uniswap_router.GetWETH(cl, ctx)
	if err != nil {
		panic(err)
	}

	return address
}

func parseFlags() {
	flagTarget := flag.String("target", "", "The hex address of the target token")
	flagEthThreshold := flag.Float64("minEthLiquidity", 30.0, "Minimum liquidity necessary to trigger the buy transaction in ETH")
	flagPrivateKeys := flag.String("privateKeys", "", "Private keys in secp256k1 format separated by coma")
	flagMaxX := flag.Int64("maxX", 3, "Maximum increase in price from listing price (multiplier e.g. 3 means 3 times the price)")
	flagOurContract := flag.String("ourContract", "", "The hex address of the smart contract that will execute the trade")

	flag.Parse()

	if common.IsHexAddress(*flagTarget) == false {
		panic("target is not a valid address")
	} else {
		target = common.HexToAddress(*flagTarget)
	}

	threshold.SetFloat64(*flagEthThreshold)

	_privateKeys := *flagPrivateKeys
	privateKeys = strings.Split(_privateKeys, ",")

	for _, p := range privateKeys {
		_, err := crypto.HexToECDSA(p)
		if err != nil {
			panic("privateKey " + p + " not valid")
		}
	}

	if common.IsHexAddress(*flagOurContract) == false {
		panic("ourContract is not a valid address")
	} else {
		ourContract = common.HexToAddress(*flagOurContract)
	}

	maxX = *flagMaxX

}
