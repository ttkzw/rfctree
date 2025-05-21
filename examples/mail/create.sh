#!/bin/sh
curdir=$(dirname "$0")

go build -o rfctree cmd/main.go
./rfctree diagram \
  --target "RFC0821,RFC0822" \
  --follow \
  --exclude-from "${curdir}/exclude.txt" \
  -output "${curdir}/MAIL-RFC.png"
