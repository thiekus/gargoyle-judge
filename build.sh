#!/bin/sh

echo Building Gargoyle Master...
# rm -f ./gymaster/*.syso
go build -v -o ./bin/gymaster ./gymaster

echo Building Gargoyle Slave...
# rm -f ./gymaster/*.syso
go build -v -o ./bin/gyslave ./gyslave
