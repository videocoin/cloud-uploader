#!/bin/bash

echo "Waiting Redis to launch on 3306..."

while ! docker run --rm --net=host gophernet/netcat -vz 127.0.0.1 6379; do
    sleep 1
done

echo "Redis launched"