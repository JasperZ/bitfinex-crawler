package influxdb

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/jasperz/bitfinex-crawler/bitfinex"
)

type config struct {
	host      string
	port      uint64
	ssl       bool
	verifySsl bool
	db        string
	username  string
	password  string
}

type InfluxDbRecorder struct {
	config *config
	client client.Client
}

func NewInfluxDbRecorder() InfluxDbRecorder {
	recorder := InfluxDbRecorder{
		config: &config{},
	}

	return recorder
}

func (r InfluxDbRecorder) ConfigFromEnv() bool {
	valid := true

	if val, set := os.LookupEnv("INFLUXDB_HOST"); set {
		r.config.host = val

		if len(val) == 0 {
			fmt.Println("INFLUXDB_HOST set but empty")
			valid = false
		}
	} else {
		fmt.Println("INFLUXDB_HOST not set, please set")
		valid = false
	}

	if val, set := os.LookupEnv("INFLUXDB_PORT"); set {
		var err error
		r.config.port, err = strconv.ParseUint(val, 10, 0)

		if err != nil {
			fmt.Printf("INFLUXDB_PORT parse error: \"%v\"\n", err)
			fmt.Println("Use default: 8086")
			r.config.port = 8086
		}
	} else {
		fmt.Println("INFLUXDB_PORT not set, use default: 8086")
		r.config.port = 8086
	}

	if val, set := os.LookupEnv("INFLUXDB_USE_SSL"); set {
		var err error
		r.config.ssl, err = strconv.ParseBool(val)

		if err != nil {
			fmt.Printf("INFLUXDB_USE_SSL parse error: \"%v\"\n", err)
			fmt.Println("Use default: false")
			r.config.ssl = false
		}
	} else {
		fmt.Println("INFLUXDB_USE_SSL not set, use default: false")
		r.config.ssl = false
	}

	if val, set := os.LookupEnv("INFLUXDB_VERIFY_SSL"); set {
		var err error
		r.config.verifySsl, err = strconv.ParseBool(val)

		if err != nil {
			fmt.Printf("INFLUXDB_VERIFY_SSL parse error: \"%v\"\n", err)
			fmt.Println("Use default: false")
			r.config.verifySsl = false
		}
	} else {
		fmt.Println("INFLUXDB_VERIFY_SSL not set, use default: false")
		r.config.verifySsl = false
	}

	if val, set := os.LookupEnv("INFLUXDB_DATABASE"); set {
		r.config.db = val

		if len(val) == 0 {
			fmt.Println("INFLUXDB_DATABASE set but empty")
			valid = false
		}
	} else {
		fmt.Println("INFLUXDB_DATABASE not set, please set")
		valid = false
	}

	if val, set := os.LookupEnv("INFLUXDB_USERNAME"); set {
		r.config.username = val

		if len(val) == 0 {
			fmt.Println("INFLUXDB_USERNAME set but empty")
			valid = false
		}
	} else {
		fmt.Println("INFLUXDB_USERNAME not set, please set")
		valid = false
	}

	if val, set := os.LookupEnv("INFLUXDB_PASSWORD"); set {
		r.config.password = val

		if len(val) == 0 {
			fmt.Println("INFLUXDB_PASSWORD set but empty")
			valid = false
		}
	} else {
		fmt.Println("INFLUXDB_PASSWORD not set, please set")
		valid = false
	}

	return valid
}

func (r InfluxDbRecorder) RecorderTask(trades <-chan bitfinex.Trade, quit <-chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	var addr string
	var clientConfig client.HTTPConfig
	var lastTradeTimestamp, tradeUniq uint64

	if r.config.ssl {
		addr = fmt.Sprintf("https://%v:%v", r.config.host, r.config.port)
	} else {
		addr = fmt.Sprintf("http://%v:%v", r.config.host, r.config.port)
	}

	clientConfig = client.HTTPConfig{
		Addr:               addr,
		InsecureSkipVerify: !r.config.verifySsl,
		Username:           r.config.username,
		Password:           r.config.password,
		Timeout:            5 * time.Second,
	}
	batchConfig := client.BatchPointsConfig{
		Database:  r.config.db,
		Precision: "ms",
	}

	// trades related variables
	lastTradeTimestamp = 0
	tradeUniq = 0

	// batch points
	batchPoints, _ := client.NewBatchPoints(batchConfig)

	writeTicker := time.Tick(10 * time.Second)

	var err error

	r.client, err = client.NewHTTPClient(clientConfig)

	if err != nil {
		fmt.Println("Recorder - Error creating InfluxDB Client: ", err.Error())
	}

	defer r.client.Close()

	for {
	mainLoop:
		for {
			select {
			case <-quit:
				fmt.Println("Recorder - Exit recorder task")
				return
			case trade := <-trades:
				if lastTradeTimestamp == trade.Timestamp {
					tradeUniq++
				} else {
					tradeUniq = 0
				}

				lastTradeTimestamp = trade.Timestamp
				batchPoints.AddPoint(r.createTradePoint(trade, tradeUniq))
			case <-writeTicker:
				if len(batchPoints.Points()) > 0 {
					err := r.client.Write(batchPoints)

					if err == nil {
						fmt.Printf("Recorder - Wrote %v data points to InfluxDB\n", len(batchPoints.Points()))
						batchPoints, _ = client.NewBatchPoints(batchConfig)
					} else {
						fmt.Println("Recorder - Error writing points to InfluxDB")
						fmt.Printf("Recorder - %v unwritten data points\n", len(batchPoints.Points()))
						break mainLoop
					}
				}
			}
		}

		fmt.Println("Recorder - Lost connection to InfluxDB")
	}
}

func (r InfluxDbRecorder) createTradePoint(trade bitfinex.Trade, uniq uint64) *client.Point {
	measurement := "trades"
	tags := map[string]string{
		"pair": trade.Pair,
		"uniq": strconv.FormatUint(uniq, 10),
	}
	fields := map[string]interface{}{
		"amount": trade.Amount,
		"price":  trade.Price,
	}
	timestamp := time.Unix(0, int64(trade.Timestamp)*int64(time.Millisecond))
	point, _ := client.NewPoint(measurement, tags, fields, timestamp)

	return point
}
