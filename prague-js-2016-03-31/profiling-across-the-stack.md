# Profiling Node.js across the stack

*Vilibald Wanča - wvi@apiary.io*

![Apiary logo](apiary-logo.png)

---

## Outline

- Before we dive in
- Basic profiling (js only)
- Going hard core (js/c++/c)

---

## Before we dive in

> What is Profiling

Measuring a resource consumption over time.

- CPU
- Memory
- io

---

## CPU Profiling methods

- *sampling*

  easy to run, easy on the system, good enough

- *instrumenting*

  precise, expensive, adds code

---

## The app is slow

Before profiling ask yourself some questions and measure.

- Is it CPU? (Event loop halt > 30ms)
- Is it memory? (GC kicking in a lot, memory growing)
- Is it IO? (Traffic)

Then choose what to do next.

---

## TALK IS CHEAP SHOW ME SOME ACTION

---

## Resources

Brendan Gregg http://www.brendangregg.com/

Strongloop https://strongloop.com/

Cpuprofilify https://github.com/thlorenz/cpuprofilify

---

## Happy profiling

> Thanks a lot.

*Vilibald Wanča (wvi@apiary.io)*

---
