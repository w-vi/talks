'use strict';

var PORT = process.env.PORT || 8000;
var http = require('http');
var server = http.createServer();

var profiler
if (process.env.V8PROFILE) {
  profiler = require('v8-profiler');
}

server
  .on('request', onRequest)
  .on('listening', onListening)
  .listen(PORT);

/// Cleanly shut down process on SIGTERM to ensure that cpuprofile is dumped
process.on('SIGTERM', onSIGTERM);

function onSIGTERM() {
  // IMPORTANT to log on stderr, to not clutter stdout which is purely for data, i.e. dtrace stacks
  console.error('Caught SIGTERM, shutting down.');
  server.close();

  if (profiler) {
    var cpuprofile = profiler.stopProfiling('fibonacci');
    require('fs').writeFileSync(
        __dirname + '/fibonacci.cpuprofile'
      , JSON.stringify(cpuprofile, null, 2)
      , 'utf8'
    );
  }

  process.exit(0);
}

console.error('pid', process.pid);

function onRequest(req, res) {
  res.writeHead(200, { 'Content-Type': 'text/plain' });

  if (req.url === '/start') {
    profiler.startProfiling('fibonacci');
    return res.end('Profiler started.\r\n');
  }

  var n = parseInt(req.url.slice(1));
  if (isNaN(n) || n < 0) return res.end('Please supply a number larger than 0, i.e. curl localhost:8000/12');

  var fib = calculateFibonacci(n);
  res.end('fibonacci(' + n + ') is ' + fib + '\r\n');
}

function onListening() {
  console.error('HTTP server listening on port', PORT);
}

function calculateFibonacci(n) {
  if (n < 2){
    return 1;
  } else {
    return calculateFibonacci(n-2) + calculateFibonacci(n-1);
  }
}
