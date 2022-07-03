# AWS Lambda using node.js (S01E01-HDTV.mkv)

---

## Outline

- What is AWS Lambda
- AWS Lambda using node.js
- Building native addons

*Ask questions straight away, don't wait for Q&A*

---

## What is AWS Lambda

> AWS Lambda is a compute service where you can upload your code
> and the service can run the code on your behalf


---

## AWS Lambda basics

- event driven
- stateless
- you manage the code not the resources
- versioning
- rollback
- stages

> ideally these are pure functions in FP sense

---

## AWS Lambda event sources

> Anything Amazon really

- API Gateway for HTTP(s)
- S3 events
- Kinesis
- DynamoDB
- CloudWatch
- ...

---

## AWS Lambda caveats and limits

- stateless
- no control when and how it is spawned
- limited by provided resources
- maximum running time 300s
- disk capacity 512 MB (1024 fds)
- Payload size max 6MB

---

##  How it works

> Linux containers advanced level

- Lambda is a container like docker
- You provide a zip file which is a overlay on a base provided by
  Amazon
- These containers are shortlived and possibly spawned many
  times concurently as needed

---

## AWS Lambda using node.js

- recommended enviroment is node.js 4.3.2
- possibly with more edge cases and issues 0.10.36

> 0.10.36 does not finish the tick of the loop


---

## The Function

Minimal example (index.js)

```js

    exports.myHandler = function(event, context, callback) {
        console.log(JSON.stringify(event));
        callback(null, "OK");
    }
```

---

## Event object

Event object is whatever the source of event sends.

API Gateway sends by default body of the post or using the Models you
can create one from Headers or body.

**This is in your hands**

---

## Context object

Context object is provided by Lambda.

*Methods*:
- `context.getRemainingTimeInMillis()` - Returns the approximate remaining execution time
- `context.succeed(obj)` - Ends function with success, obj is result
- `context.fail(err)` - Ends function with error
- `context.done(err, res)` - Ends function like a callback.

---

## Context continued

*Properties*:

- `callbackWaitsForEmptyEventLoop` - Wait for emtpy loop.
- `functionName` - name of the function
- `functionVersion` - function version
- `invokedFunctionArn` - The ARN used to invoke
- `memoryLimitInMB` - Memory limit, in MB
- `awsRequestId` - AWS request ID
- `logGroupName` - CloudWatch log group

---

## callback

- Behaves like `context.done()`
- it takes usual form `callback(err, res)`
- if `err` i not null than `res` is ignored.

---

## Building native add ons

- Must match AWS Lambda libc and family
- You can't make it locally unless running the same processor and
  Linux enviroment

*Solution?*

**DOCKER**

---

## Docker the hero

> https://github.com/lambci/docker-lambda

---

## Examples

Building

```sh
docker run -e NPM_TOKEN=${NPM_TOKEN} -v $(pwd):/var/task \
lambci/lambda:build npm install --production
```

Testing

```sh
docker run --net host -v $(pwd):/var/task lambci/lambda index.handler '{                  ⏎
  "input_type": "text/vnd.apiblueprint",
  "input_document": "## test [/test/{id}[2]]\nA test uri template",
  "output_type": "application/vnd.refract.parse-result",
  "options": {
    "source_map": false
  }
}'
```

---

## To be continued

 ... in your favourite torrent search engine soon


---
