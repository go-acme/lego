#!/usr/bin/env bash

if [[ "$1" == "timeout" ]]; then
   echo '{"timeout": 30, "interval": 5}'
else
   echo 'OOPS'
fi
