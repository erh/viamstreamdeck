#!/bin/bash
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    sudo apt update && sudo apt install -y libudev-dev
fi
