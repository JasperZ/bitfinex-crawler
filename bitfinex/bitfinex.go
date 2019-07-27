package bitfinex

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

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
	trades             chan<- Trade
}

type Trade struct {
	ID        uint64
	Timestamp uint64
	Pair      string
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

func (b BitfinexCrawler) CrawlerTask(trades chan<- Trade, quit <-chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	compiledJsonMsgRegex := regexp.MustCompile(jsonMsgRegex)
	compiledUpdateMsgRegex := regexp.MustCompile(updateMsgRegex)

	b.trades = trades

	for {
		var err error

		b.ws, _, err = websocket.DefaultDialer.Dial(apiEndpoint, nil)

		if err != nil {
			fmt.Println("Crawler - Couldn't establish Connection, reconnect in 10 seconds")
			time.Sleep(10 * time.Second)
			continue
		} else {
			fmt.Println("Crawler - Connection established")
		}

		defer b.ws.Close()

		configPhaseEnd := time.After(5 * time.Second)

		// subscribe to symbols
		for _, sym := range b.config.symbols {
			if err = b.subscribeTrade(sym); err != nil {
				break
			}
		}

		if err != nil {
			continue
		}

	mainLoop:
		for {
			select {
			case <-configPhaseEnd:
				if b.checkFinalConfiguration() {
					fmt.Println("Crawler - Configuration was successful")
				} else {
					fmt.Println("Crawler - Configuration wasn't successful")
					break mainLoop
				}
			case <-quit:
				fmt.Println("Crawler - Exit crawler task")
				return
			default:
				if _, message, err := b.ws.ReadMessage(); err == nil {
					if compiledJsonMsgRegex.Match(message) {
						b.handleJsonMessage(message)
					}

					if compiledUpdateMsgRegex.Match(message) {
						b.handleUpdateMessage(message)
					}
				} else {
					fmt.Println("Crawler - Error receiving:", err)
					break mainLoop
				}
			}
		}

		fmt.Println("Crawler - Connection closed")
		b.ws.Close()

		// reset trades subscriptions
		b.tradesChanIdSymbol = make(map[uint]string)
	}
}

func (b BitfinexCrawler) subscribeTrade(symbol string) error {
	var request = map[string]string{
		"event":   "subscribe",
		"channel": "trades",
		"symbol":  symbol,
	}

	return b.ws.WriteJSON(request)
}

func (b BitfinexCrawler) checkFinalConfiguration() bool {
	var tradesSymbolToChannel = make(map[string]uint)
	var tradesOk = true

	for k, v := range b.tradesChanIdSymbol {
		tradesSymbolToChannel[v] = k
	}

	for _, v := range b.config.symbols {
		if _, ok := tradesSymbolToChannel[v]; !ok {
			fmt.Printf("Crawler - Trade subscription for symbol %v missing\n", v)
			tradesOk = false
		}
	}

	if tradesOk {
		return true
	} else {
		return false
	}

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
		fmt.Printf("Crawler - No handler - %v\n", string(message))
	}
}

func (b BitfinexCrawler) handleUpdateMessage(message []byte) {
	compiledRegex := regexp.MustCompile(`\[(\d+),.*\]`)
	chanIdS := compiledRegex.FindStringSubmatch(string(message))[1]
	chanId, _ := strconv.ParseUint(chanIdS, 10, 64)

	for k, v := range b.tradesChanIdSymbol {
		if k == uint(chanId) {
			b.handleTradesUpdateMessage(message, v)
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
	fmt.Printf("Crawler - Subscription added - symbol: %v, chanId: %v\n", resp.Symbol, resp.ChanId)
}

func (b BitfinexCrawler) handleTradesUpdateMessage(message []byte, symbol string) {
	compiledRegex := regexp.MustCompile(`\[(\d+),("tu"),\[(\d+),(\d+),(.+),(.+)\]\]`)

	if compiledRegex.Match(message) {
		submatches := compiledRegex.FindStringSubmatch(string(message))

		id, _ := strconv.ParseUint(submatches[3], 10, 64)
		timestamp, _ := strconv.ParseUint(submatches[4], 10, 64)
		amount, _ := strconv.ParseFloat(submatches[5], 64)
		price, _ := strconv.ParseFloat(submatches[6], 64)

		trade := Trade{
			ID:        id,
			Timestamp: timestamp,
			Pair:      symbol,
			Amount:    amount,
			Price:     price,
		}

		b.trades <- trade
	}
}
