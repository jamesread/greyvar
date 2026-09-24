#!/bin/bash

[[ $_ != $0 ]] && echo "Setting CC and CXX to clang" || echo "You should not execute this script, instead, source it."

export CC=/usr/bin/clang
export CXX=/usr/bin/clang++
