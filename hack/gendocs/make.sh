#!/usr/bin/env bash

pushd $GOPATH/src/github.com/appscode/krank/hack/gendocs
go run main.go
popd
