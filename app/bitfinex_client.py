import logging
import websockets
import json
import time
import hmac
import hashlib

class BitfinexClient:
    SUPPORTED_API_VERSION = 2
    WEBSOCKET_API_URI = 'wss://api.bitfinex.com/ws/2'

    def __init__(self, symbols, tradeBuffer):
        self.logger = logging.getLogger('BitfinexClient')
        self.symbols = symbols
        self.tradeBuffer = tradeBuffer

    async def fetch(self):
        while True:
            websocket = await self._connect()

            if websocket != None:
                configured = await self._configure(websocket)

                if configured:
                    subscriptions = await self._subscribe(websocket)

                    if len(subscriptions) == len(self.symbols):
                        await self._fetchTrades(websocket, subscriptions)
            else:
                time.sleep(5)

            self._disconnect(websocket)

    async def _connect(self):
        try:
            websocket = await websockets.connect(self.WEBSOCKET_API_URI, ping_interval=5)
            self.logger.info('Connection established')
        except:
            self.logger.info('Failed to connect')
            return None
        else:
            try:
                jsonResponse = await websocket.recv()
            except:
                self.logger.info("Failed to receive data ")
                return None
            else:
                responseDict = json.loads(jsonResponse)
                serverAPIVersion = responseDict['version']

                # check for correct api version
                if not self._checkAPIVersion(serverAPIVersion):
                    self._disconnect(websocket)
                    return None
                else:
                    return websocket

    def _checkAPIVersion(self, serverAPIVersion):
        if serverAPIVersion == self.SUPPORTED_API_VERSION:
            return True
        else:
            self.logger.info('Requested API version {} is not compatible with server API version {}'
                .format(self.SUPPORTED_API_VERSION, serverAPIVersion))
            return False

    def _disconnect(self, websocket):
        if websocket != None:
            if websocket.open:
                websocket.close()
                self.logger.info('Connection closed')
            else:
                self.logger.info('Connection already closed')
        else:
            self.logger.info('No active websocket')

    async def _configure(self, websocket):
        jsonConfig = self._buildConfiguration()

        try:
            await websocket.send(jsonConfig)
        except:
            self.logger.info('Configuration failed ')
            return False
        else:
            try:
                jsonResponse = await websocket.recv()
            except:
                self.logger.info("Failed to receive data ")
                return False
            else:
                responseDict = json.loads(jsonResponse)

                if str(responseDict['event']) == 'conf' and str(responseDict['status']) == 'OK':
                    self.logger.info('Configuration successfully')
                    return True
                else:
                    self.logger.info('Configuration failed ')
                    return False

    def _buildConfiguration(self):
        flags = 32
        config = {
            "event": "conf",
            "flags": flags
        }

        return json.dumps(config)

    async def _subscribe(self, websocket):
        subscriptions = {}

        for symbol in self.symbols:
            channelId = await self._subscribeSymbol(websocket, symbol)

            if channelId != None:
                subscriptions[channelId] = symbol

        return subscriptions

    async def _subscribeSymbol(self, websocket, symbol):
        request = {
            'event': 'subscribe',
            'channel': 'trades',
            'symbol': symbol
        }
        jsonRequest = json.dumps(request)

        try:
            await websocket.send(jsonRequest)
        except:
            return None
        else:
            try:
                responseDict = {}
                event = None

                while True:
                    jsonResponse = await websocket.recv()
                    responseDict = json.loads(jsonResponse)

                    if 'event' in responseDict:
                        event = responseDict['event']

                        if event == 'subscribed' or event == 'error':
                            break
            except:
                return None
            else:
                if event == 'subscribed':
                    channelId = responseDict['chanId']
                    self.logger.info('Subscribed to symbol {}'.format(symbol))
                    return channelId
                elif event == 'error':
                    msg = responseDict['msg']
                    code = responseDict['code']
                    self.logger.error('msg: {} code: {}'.format(msg, code))
                    return None

    async def _fetchTrades(self, websocket, subscriptions):
        while True:
            try:
                jsonResponse = await websocket.recv()
            except:
                break
            else:
                responseDict = json.loads(jsonResponse)

                if 'te' not in responseDict:
                    continue

                channelId = responseDict[0]
                timestamp = responseDict[2][1]
                amount = responseDict[2][2]
                price = responseDict[2][3]

                if channelId in subscriptions:
                    await self.tradeBuffer.addTrade(subscriptions[channelId], amount, price, timestamp)
