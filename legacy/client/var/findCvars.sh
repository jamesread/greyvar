#!/bin/bash

find src \( -iname '*.cpp' -o -iname '*.hpp' \) -exec sed -rn 's/.*cvarGet[bi]?\("(.+)"\).*/\1/p' {} \; | uniq | sort
