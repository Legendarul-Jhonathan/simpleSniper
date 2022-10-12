package ws

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

type WsResponse struct {
	Jsonrpc string
	Method  string
	Params  WsParams
}

type WsParams struct {
	Subscription string
	Result       string
}

type WsResponseBlock struct {
	Jsonrpc string
	Method  string
	Params  WsParamsBlock
}

type WsParamsBlock struct {
	Subscription string
	Result       WsResultBlock
}

type WsResultBlock struct {
	Sha3Uncles       string
	Miner            string
	StateRoot        string
	TransactionsRoot string
	ReceiptsRoot     string
	LogsBloom        string
	Difficulty       string
	Number           string
	GasLimit         string
	GasUsed          string
	Timestamp        string
	ExtraData        string
	MixHash          string
	Nonce            string
	Hash             string
}

var addr string

// SetupCloseHandler creates a 'listener' on a new goroutine which will notify the
// program if it receives an interrupt from the OS. We then handle this by calling
// our clean up procedure and exiting the program.
func SetupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		os.Exit(0)
	}()
}

func NewHead(newHeadChan chan WsResponseBlock) {
	SetupCloseHandler()

	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	//u := url.URL{Scheme: "ws", Host: addr, Path: ""}
	// log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	c.WriteMessage(websocket.TextMessage, []byte(`{"id": 1, "method": "eth_subscribe", "params": ["newHeads"]}`)) //TODO: make a struct dont't hardcode a string
	// fmt.Println(errx)
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			// log.Printf("recv: %s", message)
			var res WsResponseBlock
			json.Unmarshal(message, &res)
			newHeadChan <- res

		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}

}

func NewPoolTransaction(newHeadChan chan WsResponse) {
	SetupCloseHandler()

	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	//u := url.URL{Scheme: "ws", Host: addr, Path: ""}
	// log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	c.WriteMessage(websocket.TextMessage, []byte(`{"id": 1, "method": "eth_subscribe", "params": ["newPendingTransactions"]}`)) //TODO: make a struct dont't hardcode a string
	// fmt.Println(errx)
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			// log.Printf("recv: %s", message)
			var res WsResponse
			json.Unmarshal(message, &res)
			newHeadChan <- res

		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}

}

func Init(address string) {
	addr = address
}

//TODO: cleanup
// func Init(address string, newHeadChan chan WsResponse) {
// 	address = addr
// 	SetupCloseHandler()

// 	flag.Parse()
// 	log.SetFlags(0)

// 	interrupt := make(chan os.Signal, 1)
// 	signal.Notify(interrupt, os.Interrupt)

// 	u := url.URL{Scheme: "ws", Host: addr, Path: ""}
// 	// log.Printf("connecting to %s", u.String())

// 	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
// 	if err != nil {
// 		log.Fatal("dial:", err)
// 	}
// 	c.WriteMessage(websocket.TextMessage, []byte(`{"id": 1, "method": "eth_subscribe", "params": ["newHeads"]}`)) //TODO: make a struct dont't hardcode a string
// 	// fmt.Println(errx)
// 	defer c.Close()

// 	done := make(chan struct{})

// 	go func() {
// 		defer close(done)
// 		for {
// 			_, message, err := c.ReadMessage()
// 			if err != nil {
// 				log.Println("read:", err)
// 				return
// 			}

// 			// log.Printf("recv: %s", message)
// 			var res WsResponse
// 			json.Unmarshal(message, &res)
// 			newHeadChan <- res

// 		}
// 	}()

// 	ticker := time.NewTicker(time.Second)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-done:
// 			return
// 		case <-interrupt:
// 			log.Println("interrupt")

// 			// Cleanly close the connection by sending a close message and then
// 			// waiting (with timeout) for the server to close the connection.
// 			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
// 			if err != nil {
// 				log.Println("write close:", err)
// 				return
// 			}
// 			select {
// 			case <-done:
// 			case <-time.After(time.Second):
// 			}
// 			return
// 		}
// 	}

// }

func NewUniswapPair(newHeadChan chan WsResponse) {
	SetupCloseHandler()

	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: addr, Path: ""}
	// log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	// c.WriteMessage(websocket.TextMessage, []byte(`{"id": 1, "method": "eth_subscribe", "params": ["logs", {"address": "0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f"]}`)) //TODO: make a struct dont't hardcode a string
	c.WriteMessage(websocket.TextMessage, []byte(`{"id": 1, "method": "eth_subscribe", "params": ["logs", {"address": "0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc"]}`)) //TODO: make a struct dont't hardcode a string
	// fmt.Println(errx)
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			// log.Printf("recv: %s", message)
			var res WsResponse
			json.Unmarshal(message, &res)
			newHeadChan <- res

		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("- write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}

}
