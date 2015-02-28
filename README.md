fate
====

fate is a framework for generating text. It's a formal split from my
work in the [Python](https://github.com/pteichman/cobe/) and
[Go](https://github.com/pteichman/go.cobe) versions of cobe.

cobe minimized program memory to run on a virtual machine with 64MB of
RAM, and to that end used an on-disk database for its language
model. That restriction seems quaint today.

This work begins with a highly performant trigram language model that
can generate replies to text and has reasonable memory usage.

Usage is straightforward: docs are at
[godoc.org](http://godoc.org/github.com/pteichman/fate). I've kept the
API surface very small to start.

You can try it out on the command line:

    $ go get github.com/pteichman/fate/cmd/fate-console
    $ fate-console <text files>

That will learn everything in the files (line by line) and set up an
interactive reply loop.
