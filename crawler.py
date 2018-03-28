#!/usr/bin/env python3

import os, sys
lib_path = os.path.abspath(os.path.join('apis'))
sys.path.append(lib_path)

import asyncio
import websockets
import json
import time

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

async def influxSaveTrades(host, port, ssl, verifySsl, username, password, database, queue):
    last_timestamp = ""
    uniq = 0
    client = None

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

        while not client or not client.write_points(influxTradeTemplate):
            print("No connection to InfluxDB")
            client = await connectInflux(host, port, ssl, verifySsl, username, password, database)

        last_timestamp = timestamp

async def connectInflux(host, port, ssl, verifySsl, username, password, database):
    print("Connect to InfluxDB")

    client = None

    while True:
        try:
            client = InfluxDBClient(host=host, port=port, ssl=ssl, verify_ssl=verifySsl, username=username, password=password)

            client.get_list_database()
            client.switch_database(database)
            break
        except:
            time.sleep(5)

    print("Connected to InfluxDB")

    return client


if __name__ == "__main__":
    bitfinexAPIKey = os.getenv('BITFINEX_API_KEY', None)
    bitfinexAPISecret = os.getenv('BITFINEX_API_SECRET', None)
    tickerSymbols = os.getenv('TICKER_SYMBOLS', None)

    if tickerSymbols:
        tickerSymbols = tickerSymbols.replace(" ", "").split(",")

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

    eventLoop = asyncio.get_event_loop()
    tradeQueue = queue = asyncio.Queue(loop=eventLoop)

    eventLoop.create_task(bitfinex.fetchTradesAndOrders(tickerSymbols, [], tradeQueue, bitfinexAPIKey, bitfinexAPISecret))
    eventLoop.create_task(influxSaveTrades(influxdbHost, influxdbPort, influxdbUseSSL, influxdbVerifySSL, influxdbUsername, influxdbPassword, influxdbDatabase, tradeQueue))

    eventLoop.run_forever()
