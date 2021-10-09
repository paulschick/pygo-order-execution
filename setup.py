from setuptools import setup

setup(
    name='pygo_order_execution',
    version='0.0.1',
    description='Binance US order execution client using Python and Golang',
    url='https://github.com/paulschick/pygo-order-execution',
    author='Paul Schick',
    author_email='paul@paulschick.dev',
    license='MIT',
    packages=['pygo'],
    install_requires=[
        'ccxt',
        'toolz'
    ],
    zip_safe=True
)
