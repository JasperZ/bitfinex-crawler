package bitfinex

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

const apiEndpoint string = "wss://api-pub.bitfinex.com/ws/2"
const updateMsgRegex string = "\\[([0-9]+),.*\\]"
const jsonMsgRegex string = "{\"event\":.+}"

type config struct {
	symbols []string
}

type BitfinexCrawler struct {
	config             *config
	ws                 *websocket.Conn
	tradesChanIdSymbol map[uint]string
}

type Trade struct {
	ID        uint64
	Timestamp uint64
	Amount    float64
	Price     float64
}

func NewBitfinexCrawler() BitfinexCrawler {
	crawler := BitfinexCrawler{
		config:             &config{},
		ws:                 nil,
		tradesChanIdSymbol: map[uint]string{},
	}

	return crawler
}

func (b BitfinexCrawler) ConfigFromEnv() bool {
	valid := true

	if val, set := os.LookupEnv("TICKER_SYMBOLS"); set {
		valTrimmed := strings.Trim(val, " ")
		// TODO: make sure every symbol is in the list once
		b.config.symbols = strings.Split(valTrimmed, ",")
	} else {
		fmt.Println("TICKER_SYMBOLS not set, please set")
		valid = false
	}

	return valid
}

func (b BitfinexCrawler) CrawlerTask(quit <-chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	compiledJsonMsgRegex := regexp.MustCompile(jsonMsgRegex)
	compiledUpdateMsgRegex := regexp.MustCompile(updateMsgRegex)

	for {
		var err error

		b.ws, _, err = websocket.DefaultDialer.Dial(apiEndpoint, nil)

		if err != nil {
			log.Fatal("dial:", err)
		}

		defer b.ws.Close()

		// subscribe to symbols
		for _, sym := range b.config.symbols {
			if err = b.subscribeSymbol(sym); err != nil {
				break
			}
		}

		if err != nil {
			continue
		}

		// main loop
		for {
			select {
			case <-quit:
				fmt.Println("Exit go routine")
				return
			default:
				if _, message, err := b.ws.ReadMessage(); err == nil {
					if compiledJsonMsgRegex.Match(message) {
						b.handleJsonMessage(message)
					}

					if compiledUpdateMsgRegex.Match(message) {
						b.handleUpdateMessage(message)
					}
				}
			}
		}
	}
}

func (b BitfinexCrawler) subscribeSymbol(symbol string) error {
	var request = map[string]string{
		"event":   "subscribe",
		"channel": "trades",
		"symbol":  symbol,
	}
	var jsonReq, _ = json.Marshal(request)

	return b.ws.WriteMessage(websocket.TextMessage, jsonReq)
}

func (b BitfinexCrawler) handleJsonMessage(message []byte) {
	type response struct {
		Event   string
		Channel string
		ChanId  int
	}
	var resp response

	json.Unmarshal(message, &resp)

	switch resp.Channel {
	case "trades":
		b.handleTradesJsonMessage(message)
	default:
		fmt.Printf("No handler - %v\n", string(message))
	}
}

func (b BitfinexCrawler) handleUpdateMessage(message []byte) {
	compiledRegex := regexp.MustCompile(`\[(\d+),.*\]`)
	chanIdS := compiledRegex.FindStringSubmatch(string(message))[1]
	chanId, _ := strconv.ParseUint(chanIdS, 10, 64)

	for k, _ := range b.tradesChanIdSymbol {
		if k == uint(chanId) {
			b.handleTradesUpdateMessage(message)
			break
		}
	}
}

func (b BitfinexCrawler) handleTradesJsonMessage(message []byte) {
	type response struct {
		Event   string
		Channel string
		ChanId  uint
		Symbol  string
		Pair    string
	}
	var resp response

	json.Unmarshal(message, &resp)
	b.tradesChanIdSymbol[resp.ChanId] = resp.Symbol
	fmt.Printf("Subscription added - symbol: %v, chanId: %v\n", resp.Symbol, resp.ChanId)
}

func (b BitfinexCrawler) handleTradesUpdateMessage(message []byte) {
	compiledRegex := regexp.MustCompile(`\[(\d+),("tu"),\[(\d+),(\d+),(-\d+.\d+|\d+.\d+),(\d+.\d*)\]\]`)

	if compiledRegex.Match(message) {
		submatches := compiledRegex.FindStringSubmatch(string(message))

		id, _ := strconv.ParseUint(submatches[3], 10, 64)
		timestamp, _ := strconv.ParseUint(submatches[4], 10, 64)
		amount, _ := strconv.ParseFloat(submatches[3], 64)
		price, _ := strconv.ParseFloat(submatches[3], 64)

		trade := Trade{
			ID:        id,
			Timestamp: timestamp,
			Amount:    amount,
			Price:     price,
		}

		fmt.Println(trade)
	}
}
