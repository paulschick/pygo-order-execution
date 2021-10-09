from enum import Enum


class AmountCalculation(Enum):
    FIXED_BASE_FROM_PERCENTAGE = 0
    FIXED_QUOTE_FROM_PERCENTAGE = 1
    FIXED_BASE = 2
    FIXED_QUOTE = 3


class OrderSide(Enum):
    BUY = 0
    SELL = 1
