# Channels are mutexes in disguise

*Vilibald Wanča - vilibald@wvi.cz*

---

## About me

- Double digit years of experience in the business
- Pascal -> asm x86 -> C/C++ -> Python, Lisp, Go
- Did SaaS before it was cool (2002)

*Currently I am a Bee in Apiary (apiary.io)*

---

## Outline

- What are channels anyway
- Let's see the code
- Axioms of Go channels
- When to use them?
- Performance

*Ask questions straight away, don't wait for Q&A*

---

## What are channels anyway

> Means of communication and synchronization

### Type of channels
1. "memory barrier"
2. producer-consumer queue
3. semaphore

---

## "Memory barrier"

> Synchronized unbuffered channel

*goroutine synchronously hands off the data*

    c := make(chan int)

- blocks goroutines until both are ready

---

## Producer-Consumer Queue

> Asynchronous buffered channel

*goroutine hands off data to a channel ring buffer*

    c := make(chan int, 10)

- producer blocks when buffer is full
- consumer blocks when buffer is empty
- first come first serve

---

## Semaphore

> Asynchronous buffered channel with no data

    c := make(chan struct{}, 5)

Same behaviour as in producer consumer case just less memory and more performance.

---

## Let's see the Dark side

*What is seen cannot be unseen*

- `unsafe.Pointer`
- `mallocgc`
- `atomic`

[`go/src/runtime/chan.go`](https://github.com/golang/go/blob/7e460e70d90295cf08ea627c0a0fff170aba5518/src/runtime/chan.go)

---

## Axiom #1

**A send to a nil channel blocks forever.**

```go
package main

func main() {
        var c chan string
        c <- "let's go" // deadlock
}
```

### Why

There is no space to store value, nor other side to receive value.
`*hchan` is nil `hchan` is not allocated.

---

## Axiom #2

**A send to a closed channel panics.**

```go
func main() {
    var c = make(chan int, 100)
    for i := 0; i < 10; i++ {
        go func() {
            for j := 0; j < 10; j++ {
                c <- j
            }
            close(c)
        }()
    }
    for i := range c {
        fmt.Println(i)
    }
}

```

---

## Why

The only use of channel close is to signal to the reader that there
are no more values to come.

What if function `isClosed` existed?

```go
if !isClosed(c) {
        // c isn't closed, send the value
        c <- v
}

```

*Any problems in this code?*

---

**YES**

> There is a race condition.

> What if somebody else closes `c` after you
> checked it but before you send the value?

---

## Axiom #3

**A receive from a nil channel blocks forever**

```go
package main

func main() {
        var c chan bool
        <- c // deadlock
}
```

## Why

Same as in the send case.

---

## Implications - the Issue

```go
// WaitMany waits for a and b to close.
func WaitMany(a, b chan bool) {
    var aclosed, bclosed bool
    for !aclosed || !bclosed {
        select {
            case <-a:
                aclosed = true
            case <-b:
                bclosed = true
        }
    }
}
```
Reading from close channel is always ready, see next axiom. So if `a`
is closed `bclosed` will never be set to `true`

---

## Implications - the Solution

```go
// WaitMany waits for a and b to close.
func WaitMany(a, b chan bool) {
    for a != nil || b != nil {
        select {
            case <-a:
                a = nil
            case <-b:
                b = nil
        }
    }
}
```

`select` is non-blocking so when `a` is `nil` it is skipped.

---

## Axiom #4

**A receive from a closed channel returns the value immediately**

## Why?

You can always receive from a channel because it might be buffered an
there are still values which you might want to drain and some of them
might be `zero` values.

---

## The closed indicator

You get a `bool` indication if it is closed so you can use a range
statement.

```go
for v := range c {
            // do something with v
}
```
is equal to

```go
for v, ok := <- c; ok ; v, ok = <- c {
            // do something with v
}

```

---

## Implications

```go
func main() {
        finish := make(chan struct{})
        var done sync.WaitGroup
        done.Add(1)
        go func() {
                select {
                case <-time.After(1 * time.Hour):
                case <-finish:
                }
                done.Done()
        }()
        t0 := time.Now()
        close(finish)
        done.Wait()
        fmt.Printf("Waited %v for goroutine to stop\n", time.Since(t0))
}
```

---

## Axioms of Go Channels

1. A send to a nil channel blocks forever
2. A send to a closed channel panics
3. A receive from a nil channel blocks forever
4. A receive from a closed channel returns the zero value immediately

---

## Channel vs Mutex ?

Use whichever is most expressive and/or most simple.

*Channel*
- passing ownership of data
- distributing units of work
- communicating async results

*Mutex*
- caches
- changing state

---

## Performance of the channels

- There are locks involved.
- There is scheduling involved.
- Good for most.

> If it's too slow for you?

- Try to send bigger chunks of data, less locking per item.
- Do it yourself with lock-free ring buffer.
- Maybe `sync.Mutex` and `map` will work better.

---

## Thanks a lot for your attention

Vilibald Wanča

[vilibald@wvi.cz]()

[wvi@apiary.io]()

