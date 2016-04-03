#!/bin/bash

## request page 3 times while profiling
ab -n 3 -c 1 http://:::8000/ && \
## kill process to have it write /tmp/perf-<pid>.map file
kill $1 && \
while kill -0 $1; do
    sleep 0.5
done
# run through cpuprofilify to resolve symbols and convert to cpuprofile
perf script | cpuprofilify perf > app.cpuprofile

