#!/bin/bash -e

#copy all the .proto file to here and gen them to go file.
cp ../../../game/game.proto ./
cp ../../../auth/auth.proto ./
cp ../../../bgsave/bgsave.proto ./
cp ../../../chat/chat.proto ./
cp ../../../rank/rankserver.proto ./
cp ../../../geoip/geoip.proto ./
cp ../../../snowflake/snowflake.proto ./
cp ../../../wordfilter/wordfilter.proto ./

protoc  ./*.proto --go_out=plugins=grpc:./
