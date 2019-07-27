package main

import (
	"log"
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
	trades := make(chan bitfinex.Trade, 1024)
	bitfinexConfValid := crawler.ConfigFromEnv()
	influxdbConfValid := recorder.ConfigFromEnv()

	if !bitfinexConfValid {
		log.Println("Main -", "Bitfinex config is not valid!")
	}

	if !influxdbConfValid {
		log.Println("Main -", "InfluxDB config is not valid!")
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

	<-sigint
	go func() {
		quitCrawler <- true
	}()
	go func() {
		quitRecorder <- true
	}()

	wg.Wait()
	os.Exit(0)
}
