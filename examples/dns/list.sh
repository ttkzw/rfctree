#!/bin/sh
curdir=$(dirname "$0")

go build -o rfctree cmd/main.go
./rfctree list \
  --target "RFC1034,RFC1035" \
  --follow \
  --exclude-from "${curdir}/exclude.txt"

#  --follow \
#  --keyword-from "${curdir}/keyword.txt" \

# ./rfctree list \
#   -k DNS -k DOMAIN -k 'DOMAIN NAME SYSTEM' -k 'DOMAIN NAME SPACE' \
#   -k DNSSEC -k 'DNS SECURITY' -k 'DNS-SECEXT' \
#   -k RR -k 'RESOURCE RECORD' -k 'RESOURCE RECORDS' -k 'ZONE'
