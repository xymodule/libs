#!/bin/sh
curl -L http://127.0.0.1:2379/v2/keys/backends/names -XPUT --data-urlencode value@names.txt
curl http://127.0.0.1:2379/v2/keys/seqs/USERID -XPUT -d value="0"


