#!/bin/sh

go build -o rfctree cmd/main.go
./rfctree diagram \
  --target "RFC1034,RFC1035" \
  --exclude-from exclude.txt \
  --output ./examples/dns/DNS-RFC.png

#   -k DNS -k DOMAIN -k 'DOMAIN NAME SYSTEM' -k 'DOMAIN NAME SPACE' \
#   -k DNSSEC -k 'DNS SECURITY' -k 'DNS-SECEXT' \
#   -k RR -k 'RESOURCE RECORD' -k 'RESOURCE RECORDS' -k 'ZONE'
