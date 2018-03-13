#!/usr/bin/env python3

import os, sys
lib_path = os.path.abspath(os.path.join('apis'))
sys.path.append(lib_path)

import asyncio
import websockets
import json

import bitfinex

from influxdb import InfluxDBClient

influxTradeTemplate = [
    {
        "measurement": "trades",
        "tags": {
            "pair": "",
            "uniq": 0
        },
        "time": "",
        "fields": {
            "amount": 0.0,
            "price": 0.0,
        }
    }
]

async def influxSaveTrades(client, queue):
    last_timestamp = ""
    uniq = 0

    while True:
        item = await queue.get()

        symbol = item[0]
        timestamp = item[1]
        amount = item[2]
        price = item[3]

        if timestamp == last_timestamp:
            uniq += 1
        else:
            uniq = 0

        influxTradeTemplate[0]['time'] = timestamp
        influxTradeTemplate[0]['tags']['pair'] = symbol
        influxTradeTemplate[0]['tags']['uniq'] = uniq
        influxTradeTemplate[0]['fields']['amount'] = amount
        influxTradeTemplate[0]['fields']['price'] = price

        client.write_points(influxTradeTemplate)

        last_timestamp = timestamp
        last_price = price

if __name__ == "__main__":
    bitfinexAPIKey = os.getenv('BITFINEX_API_KEY', None)
    bitfinexAPISecret = os.getenv('BITFINEX_API_SECRET', None)
    tickerSymbols = os.getenv('TICKER_SYMBOLS', None)

    if tickerSymbols:
        tickerSymbols = tickerSymbols.split(",")

    if not (bitfinexAPIKey and bitfinexAPISecret and tickerSymbols):
        print("The non optional environment variables BITFINEX_API_KEY, BITFINEX_API_SECRET and TICKER_SYMBOLS must be set")
        exit(1)

    influxdbHost = os.getenv('INFLUXDB_HOST', None)
    influxdbPort = os.getenv('INFLUXDB_PORT', 8086)
    influxdbUseSSL = os.getenv('INFLUXDB_USE_SSL', False)
    influxdbVerifySSL = os.getenv('INFLUXDB_VERIFY_SSL', False)
    influxdbDatabase = os.getenv('INFLUXDB_DATABASE', None)
    influxdbUsername = os.getenv('INFLUXDB_USERNAME', None)
    influxdbPassword = os.getenv('INFLUXDB_PASSWORD', None)

    if not (influxdbHost and influxdbDatabase and influxdbUsername and influxdbPassword):
        print("The non optional environment variables INFLUXDB_HOST, INFLUXDB_DATABASE, INFLUXDB_USERNAME and INFLUXDB_PASSWORD must be set")
        exit(1)

    client = InfluxDBClient(host=influxdbHost, port=influxdbPort, ssl=influxdbUseSSL, verify_ssl=influxdbVerifySSL, username=influxdbUsername, password=influxdbPassword)
    client.switch_database(influxdbDatabase)

    eventLoop = asyncio.get_event_loop()
    tradeQueue = queue = asyncio.Queue(loop=eventLoop)

    eventLoop.create_task(bitfinex.fetchTradesAndOrders(tickerSymbols, [], tradeQueue, bitfinexAPIKey, bitfinexAPISecret))
    eventLoop.create_task(influxSaveTrades(client, tradeQueue))

    eventLoop.run_forever()