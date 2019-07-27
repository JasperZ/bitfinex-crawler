package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/jasperz/bitfinex-crawler/bitfinex"
	"github.com/jasperz/bitfinex-crawler/influxdb"
)

func main() {
	var wg sync.WaitGroup
	crawler := bitfinex.NewBitfinexCrawler()
	recorder := influxdb.NewInfluxDbRecorder()
	sigint := make(chan os.Signal)
	quitCrawler := make(chan bool)
	quitRecorder := make(chan bool)
	trades := make(chan bitfinex.Trade, 1000)
	bitfinexConfValid := crawler.ConfigFromEnv()
	influxdbConfValid := true //recorder.ConfigFromEnv()

	if !bitfinexConfValid {
		fmt.Println("Bitfinex config is not valid!")
	}

	if !influxdbConfValid {
		fmt.Println("InfluxDB config is not valid!")
	}

	if !(bitfinexConfValid && influxdbConfValid) {
		os.Exit(1)

	}

	signal.Notify(sigint, os.Interrupt)
	signal.Notify(sigint, syscall.SIGTERM)

	wg.Add(1)
	go crawler.CrawlerTask(trades, quitCrawler, &wg)
	wg.Add(1)
	go recorder.RecorderTask(trades, quitRecorder, &wg)

	fmt.Println("Test1")
	<-sigint
	fmt.Println("Test2")
	quitCrawler <- true
	fmt.Println("Test3")
	// quitRecorder <- true
	fmt.Println("Test4")

	wg.Wait()

	os.Exit(0)
}
