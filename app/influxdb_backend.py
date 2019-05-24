import logging

from influxdb import InfluxDBClient

class InfluxDBBackend:
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

    def __init__(self, host, port, ssl, verifySsl, username, password, database, tradeBuffer):
        self.logger = logging.getLogger('InfluxDBBackend')
        self.host = host
        self.port = port
        self.ssl = ssl
        self.verifySsl = verifySsl
        self.username = username
        self.password = password
        self.database = database
        self.tradeBuffer = tradeBuffer

    async def push(self):
        last_timestamp = ""
        uniq = 0

        client = await self._connect()

        while True:
            item = await self.tradeBuffer.getTrade()

            symbol = item[0]
            amount = item[1]
            price = item[2]
            timestamp = item[3]

            if timestamp == last_timestamp:
                uniq += 1
            else:
                uniq = 0

            self.influxTradeTemplate[0]['time'] = int(timestamp)
            self.influxTradeTemplate[0]['tags']['pair'] = str(symbol)
            self.influxTradeTemplate[0]['tags']['uniq'] = int(uniq)
            self.influxTradeTemplate[0]['fields']['amount'] = float(amount)
            self.influxTradeTemplate[0]['fields']['price'] = float(price)

            while True:
                try:
                    client.write_points(self.influxTradeTemplate, time_precision='ms')
                    break
                except:
                    self.logger.info("Failed to write data")
                    client = await self._connect()

            last_timestamp = timestamp

    async def _connect(self):
        client = None

        while True:
            try:
                client = InfluxDBClient(host=self.host, port=self.port,
                    ssl=self.ssl, verify_ssl=self.verifySsl,
                    username=self.username, password=self.password)
                self.logger.info("Connection established")
            except:
                self.logger.info('Failed to connect')
                time.sleep(5)
            else:
                try:
                    client.get_list_database()
                    client.switch_database(self.database)
                except:
                    self.logger.info('Failed to switch to database {}'.format(self.database))
                else:
                    break

        return client
