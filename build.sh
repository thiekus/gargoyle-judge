#!/usr/bin/env bash
@echo off
echo Building Gargoyle Master...
go build -v -i -o ./work/gymaster.exe ./gymaster
echo Done! Press any key to exit...