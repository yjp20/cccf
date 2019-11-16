#!/bin/sh

find ./cmd -name '*.go' -exec go build {} \;
