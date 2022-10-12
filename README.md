# Simple AMM liquidity sniping bot

Today we'll look at a rather simple and out-of-date liquidity snipping bot that I built over a year ago. Although it is no longer in use, it is still a good example of how to create a bot that can snipe liquidity on AMM exchanges. This bot works with [GETH](https://geth.ethereum.org/) forks and EVM compatible chains (Ethereum, BSC, Polygon, etc).


### Context
#### Speed
When it comes to speed of execution, there is no competition between humans and machines. A bot can execute a transaction in a fraction of a second, while a human would take at least a few seconds to do the same. This is why bots are used to snipe liquidity on AMM exchanges. The bot will monitor the transaction pool and execute a transaction as soon as liquidity is added to a pair. This is a very simple strategy that can be used to make a profit on AMM exchanges, provided that the price of the token goes up after the liquidity is added.

#### Decentralized networks
There are a few key differences when comparing centralized networks to decentralized networks. One of the most important differences, in our context, is that in decentralized and open networks, the rules of the network are known to everyone and accessible to everyone.

**Transaction ordering in a block**. Orders in a block are determined by the gas price of the transaction. The higher the gas price, the higher the priority of the transaction. If two or more transactions have the same gas price, the transaction will be ordered by arrival time relative to the validator/miner.

**The body of a transaction**. A transaction contains the following information:
 - recipient address
 - signature
 - nonce
 - value
 - gas price (depending on the chain we might have extra parameters such as gasLimit, maxPriorityFeePerGas,maxFeePerGas)

Liquidity sniping refers to backrunning a transaction that adds liquidity to a pair with a buy order. If the backrun transaction is executed successfully, the bot will buy at the initial listing price and sell at a higher price.


### Simple overview of the bot
The bot connects to a GETH node and fetches the incoming transactions, it then waits for a transaction that adds liquidity to a specific token. When the bot sees the transaction, it sends multiple transactions to the smart contract that will buy the tokens.


### Walk-through of key parts in the bot
```go 
// Connect to geth
_cl, err := ethclient.Dial(IPC_ADDRESS)
if err != nil {
    panic(err)
}
ws.Init(WS_ADDRESS)
```
We initialize a connecting to GETH websocket and IPC. There reason we use IPC is because it is marginally faster than websocket, the drawback being that the bot must be on the same machine as the GETH node. The bot will start by fetching the latest block number and then start listening for incoming transactions.


```go
//Initialize various classes
token.Init(cl, ctx)
custom_contractsV2.Init()
uniswap.Init(UNISWAP_FACTORY_ADDRESS)
uniswap_pair.Init()
uniswap_router.Init(UNISWAP_ROUTER_ADDRESS)
```
Here we initialize some classes that will help us to parse incoming transaction so we can identify the 'target' transaction.
* `token.Init(cl, ctx)` token is a class used to parse ERC20 tokens, it will help us to identify the token that was added to the pair and store some information so we do not have to fetch the token data from the chain data again.
* `custom_contractsV2.Init()` is the class we use to prepare data for the transaction that will interact with our smart contract.
* `uniswap.Init(UNISWAP_FACTORY_ADDRESS)` is the class we use to interact with the uniswap factory contract.
* `uniswap_pair.Init()` is the class we use to interact with the uniswap pair contract.
* `uniswap_router.Init(UNISWAP_ROUTER_ADDRESS)` is the class we use to interact with the uniswap router contract.

*Note: because most DEX-es are based on uniswap, we can use the same classes to interact with other dex-es such as pancakeswap, quickswap, etc.*

```go
//Create workers to send transactions from
jobs = make(chan sendTransactionJob, len(privateKeys))
jobResults = make(chan common.Hash, len(privateKeys))
for w := 1; w <= len(privateKeys); w++ {
    go sendTransactionWorker(w, jobs, jobResults)
}
```
Here we create workers that will send the transactions to the network. We create a channel for the workers to receive jobs. We create a worker for each private key that we have. This is because we want to send the transactions from different accounts and send them as fast as possible.


```go
//Wait for new transaction
NewTransactionChan := make(chan ws.WsResponse)
go ws.NewPoolTransaction(NewTransactionChan)
go waitForNewTrx(NewTransactionChan)

dontExitProgramm()
```
In this part we create a channel and set up two processes on different threads that will communicate via the created channel, one will push transaction to the channel the other will consume them.
* `NewTransactionChan` is the channel that will receive the incoming transactions.
* `ws.NewPoolTransaction(NewTransactionChan)` is the function that will listen for incoming transactions and send them to the channel.
* `waitForNewTrx(NewTransactionChan)` is the function that will parse the incoming transactions and identify the 'target' transaction.


```go
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
```
Here we wait for incoming transactions and check if the transaction has information in the 'data' field. If it does we fetch the transaction by hash and send it to the `searchTrx` function.


In the `searchTrx` function we check if the destination of the transaction is the uniswap router contract.

```go
token0, token1, reserves0, reserves1, found := uniswap_router.SearchForAddLiquidity(trx.Data(), WETH)

...

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

```
If the transaction is to the uniswap router we the use the routers ABI to decode the transaction data and check if the function is `addLiquidityETH` or `addLiquidity`. If the transaction is an add liquidity transaction we then check if the liquidity contains WETH.

After we compare the amount of liquidity added with `treshold` value that we set when starting the program and also check if the tokens is the target token by comparing the token address with the `tokenAddress` value that we set when starting the program.

```go
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
```
We use `custom_contractsV2.ComposeBuyData()` method to compose the data that we will send to our smart contract. We then send the data to the workers that will send the transactions to the network by pushing to the `jobs` channel which was created at the start of the program (the number of workers is equal to the number of private keys submitted to the bot at startup).

```go
for a := 1; a <= len(privateKeys); a++ {
    h := <-jobResults
    fmt.Println(time.Now().Format("01-02-2006 15:04:05.000000"), "\t", "Tx hash from worker:", h)
}
os.Exit(0)
```
Here we wait for the workers to send the transactions and report back with the transaction hash. We then exit the program.


### Input params
`--target` The hex address of the target token

`--minEthLiquidity` Minimum liquidity necessary to trigger the buy transaction in ETH

`--privateKeys` Private keys in secp256k1 format separated by coma

`--maxX` Maximum increase in price from listing price (multiplier e.g. 3 means 3 times the price)

`--ourContract` The hex address of the smart contract that will execute the trade

### Overview of the setup
In order fot this approach to work we need GETH nodes deployed in different locations around the globe (US, Europe, Asia, etc), on each machine we then deploy this bot and the final piece is a smart contract that will validate that the order of the transaction in the block is good (this will be accomplished by checking the token price) and execute the buy order plus other logic (like anti-anti-bot logic or sending the tokens to a trade contract).

The most important factor is this approach is the time it takes to 'see' the transaction (are your nodes connected to a web3 infrastructure provider node like Infura or Binance?) and the time necessary to for your buy transaction to reach the validator/miner(are you connected to the sentry nodes of the validator/miner?).


### Closing thoughts
Although simple, this approach has been profitable but as the space is becoming more competitive this program is no longer competitive. One of the most important part of MEV is to be fast and precise, in order to achieve this you need a good infrastructure of nodes in different geographical locations, both close to the validators(sentry nodes) and close to infrastructure providers.
Also, with the advance of services like Flashbots, this approach is not necessary anymore as they provide a better way to execute MEV by tipping the miner/validator directly.


### Glossary
- AMM - automated market maker (e.g. Uniswap, Pancakeswap, Sushiswap, Balancer, etc.)
- Backrun - a transaction that is executed after another transaction
- DEX - decentralized exchange
- Frontrun - a transaction that is executed before another transaction
- MEV - miner extractable value
- Validator - a node that validates transactions and blocks
- Sentry node - a node that connects/protects the validator to the rest of the network
