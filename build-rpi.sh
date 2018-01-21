#!/bin/bash

set -eu

# example using x-tools toolchain:
# ~/x-tools/arm-unknown-linux-gnueabi/arm-unknown-linux-gnueabi/sysroot
XCOMPILER_PATH=

# example with pcap extracted and built in home dir:
# ~/src/libpcap-1.7.4
LIBPCAP_PATH=

TOOL_PREFIX="arm-unknown-linux-gnueabi-"

GOOS=linux \
GOARCH=arm \
GOARM=6 \
go build -ldflags="-extldflags=-pie" -o bin/arm/meldec ./cmd/meldec

GOOS=linux \
GOARCH=arm \
GOARM=6 \
CGO_ENABLED=1 \
CGO_CFLAGS="-I${XCOMPILER_PATH}/usr/include -I${LIBPCAP_PATH}" \
CGO_LDFLAGS="-L${XCOMPILER_PATH}/usr/lib -L${LIBPCAP_PATH}" \
CC="${TOOL_PREFIX}cc" \
LD="${TOOL_PREFIX}ld" \
go build -ldflags="-extldflags=-pie" -o bin/arm/meldec-pcap ./cmd/meldec-pcap

