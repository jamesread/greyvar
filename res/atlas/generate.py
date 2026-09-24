#!/usr/bin/env python

import json

ret = {
        "frames": dict()
}

x = 0

for i in range(3):
    ret["frames"][i] = {
        "frame": {
            "x": x,
            "y": 0,
            "w": 16,
            "h": 16
        },
        "rotated": False,
        "trimmed": False,

    }

    x += 16


print(json.dumps(ret, indent = 4))
