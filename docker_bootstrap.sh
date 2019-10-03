#!/bin/sh

echo "Starting Gargoyle Docker container"

echo "Starting Gargoyle Master..."
/opt/gargoyle/bin/gymaster > ~/gymaster.log &

echo "Starting Gargoyle Slave..."
/opt/gargoyle/bin/gyslave > ~/gyslave.log &

echo "Loading interactive shell for maintenance..."
/bin/bash --login 