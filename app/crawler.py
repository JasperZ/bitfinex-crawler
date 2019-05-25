#!/usr/bin/env python3

import logging
import asyncio
import functools
import signal
import os

from trade_buffer import TradeBuffer
from bitfinex_client import BitfinexClient
from influxdb_backend import InfluxDBBackend

def setupLogging():
    logging.basicConfig(format="%(asctime)s - %(name)s - %(levelname)s - %(message)s", level=logging.INFO)

def signal_handler(name, loop, logger):
    logger.info('Exit Application')
    eventLoop.stop()

if __name__ == "__main__":
    setupLogging()
    logger = logging.getLogger('Main')

    tickerSymbols = os.getenv('TICKER_SYMBOLS', None)

    if tickerSymbols:
        tickerSymbols = tickerSymbols.replace(" ", "").split(",")

    influxDBHost = os.getenv('INFLUXDB_HOST', None)
    influxDBPort = os.getenv('INFLUXDB_PORT', 8086)
    influxDBUseSSL = os.getenv('INFLUXDB_USE_SSL', False)
    influxDBVerifySSL = os.getenv('INFLUXDB_VERIFY_SSL', False)
    influxDBDatabase = os.getenv('INFLUXDB_DATABASE', None)
    influxDBUsername = os.getenv('INFLUXDB_USERNAME', None)
    influxDBPassword = os.getenv('INFLUXDB_PASSWORD', None)

    # if not (influxDBHost and influxDBDatabase and influxDBUsername and influxDBPassword):
    #     print("The non optional environment variables INFLUXDB_HOST, \
    #         INFLUXDB_DATABASE, INFLUXDB_USERNAME and INFLUXDB_PASSWORD must be set")
    #     exit(1)

    eventLoop = asyncio.get_event_loop()

    tradeBuffer = TradeBuffer(eventLoop)
    bitfinex = BitfinexClient(tickerSymbols, tradeBuffer)
    influxDBBackend = InfluxDBBackend(influxDBHost, influxDBPort, influxDBUseSSL,
        influxDBVerifySSL, influxDBUsername, influxDBPassword, influxDBDatabase,
        tradeBuffer)

    for signame in (signal.SIGHUP, signal.SIGUSR1, signal.SIGINT, signal.SIGTERM):
        eventLoop.add_signal_handler(signame, functools.partial(signal_handler, name=signame, loop=eventLoop, logger=logger))

    eventLoop.create_task(tradeBuffer.monitor())
    eventLoop.create_task(bitfinex.fetch())
    # eventLoop.create_task(influxDBBackend.push())

    try:
        eventLoop.run_forever()
    except:
        eventLoop.close()

    exit(0)
