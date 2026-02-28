#!/usr/bin/env bash

# Simple DNS challenge exec solver.
# Use challtestsrv https://github.com/letsencrypt/pebble/tree/main/cmd/pebble-challtestsrv#dns-01

set -e

case "$1" in
  "present")
    echo  "Present"
    payload="{\"host\":\"$2\", \"value\":\"$3\"}"
    echo "payload=${payload}"
    curl -s -X POST -d "${payload}" localhost:8055/set-txt
    ;;
  "cleanup")
    echo  "cleanup"
    payload="{\"host\":\"$2\"}"
    echo "payload=${payload}"
    curl -s -X POST -d "${payload}" localhost:8055/clear-txt
    ;;
  *)
    echo "OOPS"
    ;;
esac
