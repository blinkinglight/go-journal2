#!/bin/bash

GO111MODULE=on

cd server
go build -o jrnl_server
mv jrnl_server ../
cd ../client
go build -o note
mv note ../
cd ../
