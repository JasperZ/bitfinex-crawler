package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/jasperz/bitfinex-crawler/bitfinex"
)

func main() {
	var wg sync.WaitGroup
	crawler := bitfinex.NewBitfinexCrawler()
	influxdbConfValid := true //crawler.ConfigFromEnv()
	// influxdbConf, influxdbConfValid := influxdb.ConfigFromEnv()
	sigint := make(chan os.Signal)
	quitCrawler := make(chan bool)
	// quitRecorder := make(chan bool)
	bitfinexConfValid := crawler.ConfigFromEnv()

	if !bitfinexConfValid {
		fmt.Println("Bitfinex config is not valid!")
	}

	// if !influxdbConfValid {
	// 	fmt.Println("InfluxDB config is not valid!")
	// }

	if !(bitfinexConfValid && influxdbConfValid) {
		os.Exit(1)
	}

	signal.Notify(sigint, os.Interrupt)
	signal.Notify(sigint, syscall.SIGTERM)

	wg.Add(1)
	go crawler.CrawlerTask(quitCrawler, &wg)
	// wg.Add(1)
	// go influxdb.RecorderTask(influxdbConf, quitRecorder, &wg)

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
