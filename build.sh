#!/bin/bash

cd server
GO111MODULE=on go build -o jrnl_server
mv jrnl_server ../
cd ../client
GO111MODULE=on go build -o note
mv note ../
cd ../
