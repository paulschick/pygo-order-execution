import ccxt
import copy
import ccxt.async_support as ccxta
from .base.exchange import TradeConfig, BaseMethods
from .base.goexecute import GoExecute


class BinanceClient:
    def __init__(self, exchange, sandbox: bool):
        self.exchange = exchange
        self.sandbox = sandbox
        self.trade_configs = {}
        self.balances = {}

    def set_trade(self, key, symbol, quantity_calculation, amount, side):
        trade_ = TradeConfig({
            'symbol': symbol,
            'quantity_calculation': quantity_calculation,
            'amount': amount,
            'side': side,
            'sandbox': self.sandbox
        })
        trade_.configure_market_data(self.exchange.markets[symbol])
        self.trade_configs[key] = {
            'config': trade_
        }

    def _set_balances(self):
        balances = self.exchange.fetch_balance()
        cleaned = BaseMethods.parse_balances(balances)
        self.balances = cleaned
        for k, v in self.trade_configs.items():
            v['config'].set_balances(copy.deepcopy(cleaned))

    def _set_trades(self, key, trades):
        if key in self.trade_configs:
            self.trade_configs[key]['config'].set_trades(trades)
        else:
            raise ValueError(f'key {key} does not exist')

    def _set_orders(self, key, orders):
        if key in self.trade_configs:
            self.trade_configs[key]['config'].set_orders(orders)
        else:
            raise ValueError(f'key {key} does not exist')

    def set_state(self, days: int = 5):
        self._set_balances()
        iso = BaseMethods.date_from_days(days)
        since = self.exchange.parse8601(iso)
        for k, v in self.trade_configs.items():
            symbol = v['config'].symbol
            orders = self.exchange.fetch_orders(symbol=symbol, since=since, params={
                'timestamp': self.exchange.msec()
            })
            trades = self.exchange.fetch_my_trades(symbol=symbol, since=since, params={
                'timestamp': self.exchange.msec()
            })
            self._set_orders(k, orders)
            self._set_trades(k, trades)

    def _go_binary(self, key):
        if key in self.trade_configs:
            return GoExecute(self.trade_configs[key]['config'])
        else:
            raise ValueError(f'Key {key} does not exist')

    def place_orderbook_order(self, key):
        binary = self._go_binary(key)
        return binary.execute_orderbook_trade()

    def place_custom_order(self, key, price):
        binary = self._go_binary(key)
        return binary.execute_custom_trade(price)


class AsyncBinanceClient:
    def __init__(self, exchange, sandbox: bool):
        self.exchange = exchange
        self.sandbox = sandbox
        self.trade_configs = {}
        self.balances = {}

    def set_trade(self, key, symbol, quantity_calculation, amount, side):
        trade_ = TradeConfig({
            'symbol': symbol,
            'quantity_calculation': quantity_calculation,
            'amount': amount,
            'side': side,
            'sandbox': self.sandbox
        })
        trade_.configure_market_data(self.exchange.markets[symbol])
        self.trade_configs[key] = {
            'config': trade_
        }

    async def _set_balances(self):
        balances = await self.exchange.fetch_balance()
        await self.exchange.close()
        cleaned = BaseMethods.parse_balances(balances)
        self.balances = cleaned
        for k, v in self.trade_configs.items():
            v['config'].set_balances(copy.deepcopy(cleaned))

    def _set_trades(self, key, trades):
        if key in self.trade_configs:
            self.trade_configs[key]['config'].set_trades(trades)
        else:
            raise ValueError(f'key {key} does not exist')

    def _set_orders(self, key, orders):
        if key in self.trade_configs:
            self.trade_configs[key]['config'].set_orders(orders)
        else:
            raise ValueError(f'key {key} does not exist')

    async def set_state(self, days: int = 5):
        try:
            await self._set_balances()
            iso = BaseMethods.date_from_days(days)
            since = self.exchange.parse8601(iso)
            for k, v in self.trade_configs.items():
                symbol = v['config'].symbol
                orders = await self.exchange.fetch_orders(symbol=symbol, since=since, params={
                    'timestamp': self.exchange.msec()
                })
                trades = await self.exchange.fetch_my_trades(symbol=symbol, since=since, params={
                    'timestamp': self.exchange.msec()
                })
                await self.exchange.close()
                self._set_orders(k, orders)
                self._set_trades(k, trades)
        except Exception as e:
            print(f'Ran into exception {type(e).__name__}\n'
                  f'{str(e)}')
            await self.exchange.close()
            raise e

    def _go_binary(self, key):
        if key in self.trade_configs:
            return GoExecute(self.trade_configs[key]['config'])
        else:
            raise ValueError(f'Key {key} does not exist')

    def place_orderbook_order(self, key):
        binary = self._go_binary(key)
        return binary.execute_orderbook_trade()

    def place_custom_order(self, key, price):
        binary = self._go_binary(key)
        return binary.execute_custom_trade(price)


class ClientConfiguration:
    exchange = None
    sandbox = False
    _binance_client = None

    @classmethod
    def configure(cls, apikey, secret, sandbox=False) -> BinanceClient:
        cls.sandbox = sandbox
        if cls.exchange is None:
            cls.exchange = ccxt.binanceus({
                'apiKey': apikey,
                'secret': secret
            })
            cls.exchange.load_markets()
            cls._binance_client = BinanceClient(cls.exchange, sandbox)
            return cls._binance_client
        else:
            return cls._binance_client


class AsyncConfiguration:
    exchange = None
    sandbox = False
    _binance_client = None

    @classmethod
    async def configure(cls, apikey, secret, sandbox=False) -> AsyncBinanceClient:
        cls.sandbox = sandbox
        if cls.exchange is None:
            cls.exchange = ccxta.binanceus({
                'apiKey': apikey,
                'secret': secret
            })
            await cls.exchange.load_markets()
            await cls.exchange.close()
            cls._binance_client = AsyncBinanceClient(cls.exchange, sandbox)
            return cls._binance_client
        else:
            return cls._binance_client
