# PyGo Order Execution - Binance US Limit Order Client

Version 0.0.1

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

- ETL logic written in Python using CCXT
- Order execution logic written in Golang and called from Python 

## Summary

This project was started for a few reasons:

1. To make it easier to place orderbook and custom trades on Binance US
2. To increase performance when it comes to executing orders

Python is used to handle data formatting and setup. It calls a Golang binary to fetch the orderbook
price and execute a limit order. This project requires that you have a `.env` file in the root of your
project, which is used by the Golang binary.

NOTE: This is an alpha testing release. Use at your own risk.
Currently, this project has only been tested on Linux. I plan on creating a docker container
for the project to improve platform compatibility.


## Usage

1. Clone the repository:
`git clone https://github.com/paulschick/pygo-order-execution`
2. Install the package with pip into a virtual environment

```shell
cd pygo-order-execution
pip install -e .
```
