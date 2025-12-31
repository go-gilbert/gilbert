#!/usr/bin/env sh
set -e
go run ../cmd/gilbert2 "$@" --log-level=debug
