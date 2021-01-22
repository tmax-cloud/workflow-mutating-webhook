#!/bin/sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-s' -o main .
#CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-s' -o main main.go
