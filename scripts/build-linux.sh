#!/bin/bash
# Builds the binary (for Linux only), adding the tag for static compilation. Without using this tag,
# the runtime will require glibc to be installed.
GOOS=linux GOARCH=amd64 go build -tags=netgo -o ./onetimesharer-linux-arm64
