#!/usr/bin/env sh
set -e
go run ../../cmd/gilbert "$@" --log-level=debug
