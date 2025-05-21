#!/bin/sh

go build -o rfctree cmd/main.go
./rfctree keyword \
  -k DNS -k DOMAIN -k 'DOMAIN NAME SYSTEM' -k 'DOMAIN NAME SPACE' \
  -k DNSSEC -k 'DNS SECURITY' -k 'DNS-SECEXT' -k 'DOMAIN NAME SYSTEM SECURITY EXTENSIONS' \
  -k RR -k 'RESOURCE RECORD' -k 'RESOURCE RECORDS' -k 'ZONE'
