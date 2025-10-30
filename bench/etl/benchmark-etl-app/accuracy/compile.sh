#!/bin/sh

TARGET="main.go types.go service.go repository.go"
GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o bootstrap ${TARGET}