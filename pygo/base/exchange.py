import os
from pathlib import Path
from datetime import datetime, timezone, timedelta
from toolz import dissoc
from typing import Dict
from .models import AmountCalculation


class BaseMethods:
    @staticmethod
    def safe_dict(key, data):
        return data[key] if key in data else None

    @staticmethod
    def utc_timestamp() -> int:
        dt = datetime.now(timezone.utc)
        return int(dt.timestamp() * 1000)

    @staticmethod
    def parse_balances(balances) -> Dict:
        return dissoc(balances, 'free', 'used', 'total', 'info', 'timestamp', 'datetime')

    @staticmethod
    def date_from_days(days: int) -> str:
        days_ts = datetime.now(timezone.utc) - timedelta(days)
        return days_ts.strftime('%Y-%m-%dT%H:%M:%S.%f%z')

    @staticmethod
    def binary_filepath():
        return os.path.join(Path(__file__).parent.parent, 'placeOrder.so')


class TradeConfig(BaseMethods):
    def __init__(self, config=None):
        if config is None:
            config = {}
        self.symbol = self.safe_dict('symbol', config)
        self.quantity_calculation: AmountCalculation = self.safe_dict('quantity_calculation', config)
        self.amount = self.safe_dict('amount', config)
        self.side = self.safe_dict('side', config)
        self.sandbox = self.safe_dict('sandbox', config)

        self.market = None
        self.base = None
        self.base_balance = None
        self.quote = None
        self.quote_balance = None
        self.orders = []
        self.trades = []

    @property
    def amount_precision(self) -> int:
        return self.market['precision']['amount'] if self.market is not None else None

    @property
    def price_precision(self) -> int:
        return self.market['precision']['price'] if self.market is not None else None

    def configure_market_data(self, market: Dict):
        self.market = market
        if self.symbol is not None:
            base, quote = self.symbol.split('/')
            self.base = base
            self.quote = quote

    def set_balances(self, balances: Dict):
        if self.symbol is not None and self.base is None and self.quote is None:
            base, quote = self.symbol.split('/')
            self.base = base
            self.quote = quote
        if self.base in balances:
            self.base_balance = balances[self.base]['free']
        if self.quote in balances:
            self.quote_balance = balances[self.quote]['free']

    def set_trades(self, trades):
        if type(trades) is list:
            self.trades = [dissoc(trade, 'info') for trade in trades]
        else:
            self.trades = [dissoc(trades, 'info')]

    def set_orders(self, orders):
        if type(orders) is list:
            self.orders = [dissoc(order, 'info') for order in orders]
        else:
            self.orders = [dissoc(orders, 'info')]

    def get_state(self):
        return {
            'symbol': self.symbol,
            'base': {
                self.base: self.base_balance
            },
            'quote': {
                self.quote: self.quote_balance
            },
            'orders': self.orders,
            'trades': self.trades,
            'precision': {
                'amount': self.amount_precision,
                'price': self.price_precision
            }
        }
