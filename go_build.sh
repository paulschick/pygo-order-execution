#!/bin/bash

parent_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" ; pwd) || exit

cd "$parent_dir"/go || exit

go build -buildmode=c-shared -o placeOrder.so placeOrder.go

dest="$parent_dir"/pygo

cp placeOrder.so "$dest"
