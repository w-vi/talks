#!/usr/bin/env bash

## get everything initialized           \
curl localhost:8000/10 &&               \
                                        \
## start profiler                       \
curl localhost:8000/start &&            \
                                        \
## make couple of resuest               \
ab -n 10 http://:::8000/30 &&           \
                                        \
## kill app and have it write profile   \
kill $1
