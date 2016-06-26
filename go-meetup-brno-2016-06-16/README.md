# Channel is Mutex in disguise

Go tries hard to help you with concurrency and it is doing pretty good
job. But, there is always some but and this talk is about it. Go
channels are great but not always and not everywhere. The talk aims to
shed light on when and how to use them by revealing the implementation.

This is a material I used for the talk I gave on a 2nd golang meetup
in Brno on 16th of June 2016.

Slides are done using
[reveal-md](https://github.com/webpro/reveal-md).

The `chan.go` file contains a transcript of sorts of the code
walk-through given on the talk. See the `docs/chan.html` for a html
version. Or process the `chan.go` file with
[gocco](http://nikhilm.github.io/gocco/).


*The title is intentionally controversial a bit but don't expect any bold statements as the reality is always more nuanced.*
