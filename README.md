### Docker ###
Pull ticker data from Bitfinex and save it to the InfluxDB time series database.

## Configuration via environment variables ##
# Bitfinex API configuration #
- `BITFINEX_API_KEY` Key to access the Bitfinex API
- `BITFINEX_API_SECRET` Secret to access the Bitfinex API
- `TICKER_SYMBOLS` Comma seperated list of ticker symbols to track, e.g. `tBTCUSD, tETHUSD`

# InfluxDB configuration #
- `INFLUXDB_HOST` Hostname to connect to InfluxDB
- `INFLUXDB_PORT` Port to connect to InfluxDB **(defaults to 8086)**
- `INFLUXDB_USE_SSL` Use https instead of http to connect to InfluxDB **(defaults to False)**
- `INFLUXDB_VERIFY_SSL` Verify SSL certificates for HTTPS requests **(defaults to False)**
- `INFLUXDB_DATABASE` Database name to connect to
- `INFLUXDB_USERNAME` Username used to connect to InfluxDB
- `INFLUXDB_PASSWORD` Password used to connect to InfluxDB