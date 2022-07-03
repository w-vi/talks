# DevOps 4 real 
### Apiary walking the walk

*Vilibald Wanča - vilibald@wvi.cz*

---

## Who the hell is this guy

Running the show in Apiary/Oracle as of now.

*Architecture, Firefighting, Infrastructure*

---

## Disclaimer

*I am going to talk about what used to be Apiary as I am not allowed to talk
about the current state.*

---

## What am I going to talk about

- What is Apiary
- Infrastructure in Apiary
- SDLC
- Ops? What Ops?

*Ask questions straight away, don't wait for Q&A*

---

## What is Apiary

Apiary is a platform designed to help companies to accelerate and control the
development of their APIs.

*Marketing alert*

---

## What is Apiary for real

![Apiary](What-is-Apiary.svg)

---

## Some facts

- ~ 400K users
- ~ 500K API projects
- 15 req/s on parsing
- 99.95% uptime SLA (~5 hrs a year)
- No scheduled downtime
- Team is ~20 devs

---

## Apiary architecture

![Infrastructure](infra.jpg)

---

> In many ways oldschool.

<p class="fragment" data-fragment-index="1">Core is a classic 3 tier monolith</p>
<p class="fragment" data-fragment-index="2">The rest is so called "cloud native"</p>

---

![Architecture](Arch.svg)

---

## How we scale it?

> Mostly classic approach

<p class="fragment" data-fragment-index="1">More application instances</p>
<p class="fragment" data-fragment-index="2">Asynchronous processing (scaling workers)</p>
<p class="fragment" data-fragment-index="3">Database sharding</p>
<p class="fragment" data-fragment-index="4">"Serverless"</p>

---

## Tech stack

- Node.js
- Redis (cache)
- MongoDB (main database)
- Little PostgreSQL
- Parsing is C/C++
- RabbitMQ
- Tools are in Go 
- Data analysis is Python

---

## SDLC

> Software development life cycle

![sdlc](sdlc.png)

---

## How we do it

- Git everything
- Code review everything
- CI all the time
- CD everywhere but production
- Trello + Github + Slack
- Automate, automate, automate
- Zapier, webhooks, bots

---

## From idea to commit

> Trello or/and Github

- Some form of agile
- Trello boards for everything (automated)
- Github issues
- Regular planning sessions

---

## From commit to prod

![Commit to Prod](commit-prod.svg)

---

## Oops! Where is Ops?

![britney](britney.jpg)

---

## No Ops but DevOps

> **Every engineer is on call***

![devops](devops.gif)

<small>* unless junior or recently joined  </small>

---

## On call 

![oncall](oncall.svg)

---

## Incident management

![bomb](bomb.gif)

---

## High Priority 

> Wake them up

- Run books
- One communication channel
- Always do a post mortem 
- Regular reviews

---

## Low Priority

> Let them sleep

- Investigate
- Fix
- Add to backlog
- Leave it as is
- Regular reviews

---

## SRE team

> The only team not doing any product work

- Running basic infrastructure
- Guidance on tech decisions
- Pre-baking solutions for logging, monitoring, alerting
- Helping other teams to operate the service(s)
- Senior SREs are Incident command

---

## This presentation is over

![byebye](byebye.gif)

