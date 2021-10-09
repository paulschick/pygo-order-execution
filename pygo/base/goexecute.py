import ctypes
from .exchange import TradeConfig, BaseMethods
from .models import AmountCalculation, OrderSide


class GoExecute:
    def __init__(self, trade_config: TradeConfig):
        self.config = trade_config
        self.amount_type = None
        self.amount = None
        self.side = self.set_side(trade_config)
        self.trade_mode = self.set_trade_mode(trade_config)
        self.set_amounts()

    def set_amounts(self):
        if self.config.quantity_calculation.value == AmountCalculation.FIXED_QUOTE.value:
            self.amount = self.config.amount
            self.amount_type = 1
        elif self.config.quantity_calculation.value == AmountCalculation.FIXED_QUOTE_FROM_PERCENTAGE.value:
            self.amount = self.config.quote_balance * self.config.amount
            self.amount_type = 1
        elif self.config.quantity_calculation.value == AmountCalculation.FIXED_BASE.value:
            self.amount = self.config.amount
            self.amount_type = 0
        elif self.config.quantity_calculation.value == AmountCalculation.FIXED_BASE_FROM_PERCENTAGE.value:
            self.amount = self.config.base_balance * self.config.amount
            self.amount_type = 0
        else:
            raise ValueError("Invalid amount/amount type")

    @staticmethod
    def set_side(config: TradeConfig):
        if config.side.value == OrderSide.BUY.value:
            return 0
        else:
            return 1

    @staticmethod
    def set_trade_mode(config: TradeConfig):
        if config.sandbox is True:
            return 1
        else:
            return 0

    @staticmethod
    def load_orderbook_binary():
        place_order_bin = ctypes.cdll.LoadLibrary(BaseMethods.binary_filepath())
        place_order = place_order_bin.PlaceOrder

        place_order.argtypes = [
            ctypes.c_char_p, ctypes.c_int, ctypes.c_int,
            ctypes.c_char_p, ctypes.c_int, ctypes.c_int
        ]
        return place_order

    @staticmethod
    def load_custom_price_binary():
        place_order_bin = ctypes.cdll.LoadLibrary(BaseMethods.binary_filepath())
        custom_price = place_order_bin.TradeCustomPrice

        custom_price.argtypes = [
            ctypes.c_char_p, ctypes.c_int, ctypes.c_int,
            ctypes.c_char_p, ctypes.c_int, ctypes.c_int,
            ctypes.c_char_p, ctypes.c_int
        ]
        return custom_price

    def execute_orderbook_trade(self):
        binary = self.load_orderbook_binary()
        symbol = f'{self.config.base}{self.config.quote}'.encode('utf-8')
        amount = f'{self.amount}'.encode('utf-8')
        return binary(symbol, self.side, self.amount_type, amount, self.config.amount_precision, self.trade_mode)

    def execute_custom_trade(self, price):
        binary = self.load_custom_price_binary()
        symbol = f'{self.config.base}{self.config.quote}'.encode('utf-8')
        amount = f'{self.amount}'.encode('utf-8')
        price_ = f'{price}'.encode('utf-8')
        return binary(symbol, self.side, self.amount_type, amount, self.config.amount_precision,
                      self.trade_mode, price_, self.config.price_precision)
