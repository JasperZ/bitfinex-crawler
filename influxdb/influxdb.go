package influxdb

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/jasperz/bitfinex-crawler/bitfinex"
)

type config struct {
	host      string
	port      uint64
	ssl       bool
	verifySsl bool
	db        string
	user      string
	password  string
}

type InfluxDbRecorder struct {
	config *config
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
		r.config.user = val

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

	for {
		fmt.Printf("Write Trade to db: %v\n", <-trades)
	}
}
