# AWS Lambda and API Gateway (S01E02-HDTV.mkv)

---

## Outline

- Deployment and Invocation
- API Gateway integration
- Tips/Tricks/Caveats/Issues

*Ask questions straight away, don't wait for Q&A*

---

## How to deploy the lambda function

- aws cli
- apex
- serverless

*I don't like serverless so I won't cover it*

---

### AWS CLI

> Very simple and fo smaller function with no native code it is enough

- create a zip file
- upload the code
- updated alias ?

*Example*
https://github.com/apiaryio/cloudwatch-to-papertrail/

---

### Apex

http://apex.run/

- simple configuration to manage project of more functions
- hooks to customize steps
- utility functions like `logs` `rollback` etc.
- custom binary support using node.js shim (`exec`)

*Example:*
https://github.com/apiaryio/helium/

---

### Apex usefull commands

- build
- deploy
- invoke
- logs

---

## API Gateway

> Fully managed service by amazon to manage HTTP APIs

- pretty straightforward usage
- supports authentication, monitoring, API Tokens  etc
- tightly coupled with lambda
- allows to have multile deployments aka stages
- is powerfull but has idiosyncracies which get in the way


---

## API GW Helium Demo

- UI
- Swagger
- Templates

---

## The funny and not so funny stuff

- Don't forget the limits
- Templates
- Managing return codes

---

### API GW Limits

- Integration timeout: 30 second (HARD)
- Payload size: 10 MB (HARD)
- Stages per API:  10 (SOFT)
- Throttling limits per account:  1000 rps and 2000 burst (SOFT)

---

### VTL (Velocity Template Langugae)

- text only
- whitespace matters
- `$util.parseJson($input.path('$.errorMessage')))`
- `$util.escapeJavaScript($err.message).replace("\'", "'")`
- `foreach` to iterate over keys

*HELIUM swagger has something to offer*
https://github.com/apiaryio/helium/blob/master/swagger.yaml

https://velocity.apache.org/engine/1.7/vtl-reference.html

---

### Error handling
*Here I come. They call me trouble.*

**By default everything is 200 OK**

> Any other code has to be explicitly done.

Regex matching the log messages

- error callback from lambda
- crashes/timeouts
- 500 produced by amazon internally

----

## To be continued

 ... in your favourite torrent search engine soon

---


