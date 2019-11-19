#!/bin/sh

mkdir -p ./build
find "$PWD/cmd" -name '*.go' -exec sh -c 'go build -o "build/$(basename {} .go)" {}' \;
