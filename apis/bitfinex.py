import asyncio
import websockets
import json
import time
import hmac
import hashlib

SUPPORTED_API_VERSION = '2'
WEBSOCKET_API_URI = 'wss://api.bitfinex.com/ws/2'

bitfinexAuthenticationRequest = {
    'event': 'auth',
    'apiKey': 'api_key',
    'authSig': 'signature',
    'authPayload': 'payload',
    'authNonce': '+authNonce',
    'calc': 1
}

bitfinexConfigurationRequest = {
    "event": "conf",
    "flags": "flags"
}

bitfinexSubscribeTradeRequest = {
    "event": "subscribe",
    "channel": "trades",
    "symbol": "symbol"
}

async def fetchTradesAndOrders(tradingSymbols, orderSymbols, tradeQueue, apiKey, apiSecret):
    while True:
        websocket = await connect()

        if websocket != None:
            if await configure(websocket):
                tradingSubscriptions = await subscribeToTradingSymbols(websocket, tradingSymbols)

                #await authenticate(websocket, apiKey, apiSecret)

                if len(tradingSubscriptions) == len(tradingSymbols):
                    await fetchTrades(websocket, tradingSubscriptions, tradeQueue)

            await disconnect(websocket)

        print()

async def connect():
    try:
        websocket = await websockets.connect(WEBSOCKET_API_URI)
    except:
        print('Connection could NOT be established')
        return None
    else:
        try:
            json_response = await websocket.recv()
            response_dict = json.loads(json_response)

            # check for correct api version
            if str(response_dict['version']) == SUPPORTED_API_VERSION:
                print('Connection successfully established')
                return websocket
        except:
            print('Connection could NOT be established, API version {} not supported'.format(response_dict['version']))
            return None
            
async def disconnect(websocket):
    if websocket != None:
        websocket.close()

        print('Connection successfully closed')
    else:
        print('Connection could NOT be closed, maybe it was not established')

async def configure(websocket):
    bitfinexConfigurationRequest['flags'] = 32

    try:
        await websocket.send(json.dumps(bitfinexConfigurationRequest))
        json_response = await websocket.recv()
        response_dict = json.loads(json_response)

        if str(response_dict['event']) == 'conf' and str(response_dict['status']) == 'OK':
            print('Connection successfully configured')
            return True
        else:
            print('Connection could NOT be configured ')
            return False
    except:
            print('Connection could NOT be configured ')
            return False

async def authenticate(websocket, apiKey, apiSecret):
    nonce = int(time.time())
    payload = 'AUTH{}'.format(nonce)
    sig = hmac.new(apiSecret.encode(), msg=payload.encode(), digestmod='sha384')
    payloadSign = sig.hexdigest()

    bitfinexAuthenticationRequest['apiKey'] = apiKey
    bitfinexAuthenticationRequest['authSig'] = payloadSign
    bitfinexAuthenticationRequest['authPayload'] = payload
    bitfinexAuthenticationRequest['authNonce'] = nonce

    print(json.dumps(bitfinexAuthenticationRequest))

    await websocket.send(json.dumps(bitfinexAuthenticationRequest))
    json_response = await websocket.recv()

    print(json_response)

async def subscribeToTradingSymbols(websocket, tradingSymbols):
    tradingSubscriptions = {}

    for symbol in tradingSymbols:
        bitfinexSubscribeTradeRequest['symbol'] = symbol

        try:
            await websocket.send(json.dumps(bitfinexSubscribeTradeRequest))
            json_response = await websocket.recv()
            response_dict = json.loads(json_response)

            if 'event' in response_dict:
                if response_dict['event'] == 'subscribed':
                    channelId = int(response_dict['chanId'])
                    tradingSubscriptions[channelId] = symbol
                    json_response = await websocket.recv()

                if response_dict['event'] == 'error':
                    print(response_dict)
                    break
        except:
            break

    subscribedTraidingPairs = []

    for subscription in tradingSubscriptions:
        subscribedTraidingPairs.append(tradingSubscriptions[subscription])

    print('successfully subscribed to the following trading pairs: {}'.format(subscribedTraidingPairs))

    return tradingSubscriptions

async def fetchTrades(websocket, subscribedSymbols, queue):
    while True:
        try:
            response = await asyncio.wait_for(websocket.recv(), timeout=1)

            if 'event' in response:
                break
        except:
            try:
                pong_waiter = await websocket.ping()
                await asyncio.wait_for(pong_waiter, timeout=1)
                #print('Connection still alive')
            except:
                print('Connection not alive')
                break
        else:
            response_splitted = response.replace('[', '').replace(']', '').replace('"', '').split(',')
            channelId = int(response_splitted[0])

            if channelId in subscribedSymbols:
                if response_splitted[1] == 'te':
                    timestamp = str(response_splitted[3])
                    amount = float(response_splitted[4])
                    price = float(response_splitted[5])

                    await queue.put((subscribedSymbols[channelId], timestamp, amount, price))
