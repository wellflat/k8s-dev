#!/bin/sh

TARGET="main.go types.go dynamodb.go"
GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o main ${TARGET}