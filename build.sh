#!/bin/sh

echo Building Gargoyle Master...
rm -f ./gymaster/*.syso
go build -v -i -ldflags="-s -w" -o ./bin/gymaster ./gymaster

echo Building Gargoyle Slave...
rm -f ./gymaster/*.syso
go build -v -i -ldflags="-s -w" -o ./bin/gyslave ./gyslave
