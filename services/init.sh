#!/bin/sh
curl -L http://127.0.0.1:4001/v2/keys/backends/names -XPUT --data-urlencode value@names.txt