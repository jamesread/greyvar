#!/bin/bash

./cmakeClean.sh
cmake -DCMAKE_TOOLCHAIN_FILE=windows.toolchain .
