# Profiling Node.js app across the stack

This is material I used for the talk I gave on Prague JS meetup.

Slides are done using [reveal-md](https://github.com/webpro/reveal-md)
but there's not much information it was more of a hands on talk.

## Demo projects
Most of the talk is about showing various approaches in profiling
node.js apps. The results are always in a `results` subdirectory so
you don't need to run it to see how it looks. For both you need at
least node v0.12.

### Demo-cpu

This is a simple server which computes the Fibonacci numbers the worst
way you can do it, recursive method which is CPU intensive but shows
itself clearly in the profiling results.

`npm run start` to get the basic profiling using the builtin `--prof`
flag. It creates a [v8.log](https://github.com/v8/v8/wiki/V8-Profiler)
file which then needs to be processed to make it useful for a
human. On node 0.12 you need to install the
[`node-tick-processor`](https://www.npmjs.com/package/tick) package on
newer node version there is a builtin flag `--prof-process` which does
the same thing.

`npm run start-v8profiler` and then `profile-v8.sh` to get the results
using the [`v8-profiler`](https://www.npmjs.com/package/v8-profiler)
package which allows you to profile only the interesting bits not the
whole app. The result is a `cpuprofile` file which can be then loaded
into
[Chrome DevTools](https://developers.google.com/web/tools/chrome-devtools/)
for inspection.


## Demo-offcpu
This is a simple server which tries to simulate slow connection to
other services. Is serves a tip of the day message which it gets over
https.

`npm run start-prof` will profile the app using the `--prof` approach
which in this case is not useful as the bottle neck is not on CPU.
`ab -n 3 -c 1 http://:::8000/` can be used as a testing command to
make the server work.

`npm run start-perf` will profile the app using linux
[perf](https://perf.wiki.kernel.org/index.php/Main_Page) tool to get
also the C/C++ stack information. To run the profiling see
`profile.sh` file. Because the perf tool produces loads of data it
needs to be processed for easier visualizations. There is a tool
called [cpuprofilify](https://www.npmjs.com/package/cpuprofilify)
which does exactly that a produces again the `cpuprofile` file.


## Credit

[Thorsten Lorenz](https://github.com/thlorenz) for inspiration for demo apps which I
have modified a bit.
[Brendand Gregg](http://www.brendangregg.com/) for the perf tool
description and ideas
