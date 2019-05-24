import logging
import asyncio

class TradeBuffer:
    def __init__(self, eventLoop):
        self.logger = logging.getLogger('TradeBuffer')
        self.queue = asyncio.Queue(loop=eventLoop)

    async def addTrade(self, symbol, amount, price, timestamp):
        await self.queue.put((symbol, amount, price, timestamp))

    async def getTrade(self):
        return await self.queue.get()

    async def monitor(self):
        while True:
            self.logger.info("queue size: {}".format(self.queue.qsize()))
            await asyncio.sleep(5)
