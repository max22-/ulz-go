#!/bin/sh

xxd -r -p encoded_data.txt > encoded_data.bin
go build
./ulz encoded_data.bin output.txt