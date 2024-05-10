#!/usr/bin/env bash
# Stops the process if something fails
set -xe

# get all of the dependencies needed
go get

# create the application binary that eb uses
GOOS=linux GOARCH=amd64 go build -o bin/application -ldflags="-s -w"