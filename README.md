`rmq` is a command line Swiss army knife for sending and receiving messages to and from RabbitMQ.

[![Build Status](https://travis-ci.org/relops/rmq.png?branch=master)](https://travis-ci.org/relops/rmq)
[![Download](https://api.bintray.com/packages/relops/rmq/rmq/images/download.png)](https://bintray.com/relops/rmq/rmq/_latestVersion)

Example
-------

To send a random message to a queue:

```
$ rmq -d in -c 1 -k foo
2014-12-03 10:53:59.812 - sender connected to localhost
2014-12-03 10:53:59.812 - [290980845845254145] sending 64 bytes (8bd39598)
```

To receive messages from a queue:

```
$ rmq -d out -q foo
2014-12-03 10:53:57.024 - receiver connected to localhost
2014-12-03 10:53:57.026 - receiver subscribed to queue: foo
2014-12-03 10:53:59.813 - [290980845845254145] receiving 64 bytes (8bd39598) @ 0.81 ms
```

Features
--------

* Send and receive messages to RabbitMQ from the command line
* Send an arbitrary number of messages
* Specify the average size and standard deviation of the messages to send
* Crude send rate throttling
* Use persistent messaging as an option
* Prints latency metrics for round trip operations
* Deep entropy analysis for sending and receiving messages
* Optionally auto-re-subscribe to cancelled subscriptions (e.g. with mirrored queues)

Installation
------------

On OSX you can use Homebrew to install rmq:

```
$ brew tap relops/homebrew-rmq
$ brew install rmq
```

On Linux and OSX, you can download the binary: [![Download](https://api.bintray.com/packages/relops/rmq/rmq/images/download.png)](https://bintray.com/relops/rmq/rmq/_latestVersion)

If your platform is not covered here, please get in touch and we can probably cross-compile it for you.

Options
-------

```
$ rmq -h
Usage:
  rmq [OPTIONS]

Application Options:
  -d, --direction=  Use rmq to send (-d in) or receive (-d out) messages
  -x, --exchange=   The exchange to send to (-d in) or bind a queue to when receiving (-d out)
  -q, --queue=      The queue to receive from (when used with -d in)
  -P, --persistent  Use persistent messaging (false)
  -n, --no-declare  If set, then don't attempt to declare the queue or bind it (false)
  -k, --key=        The key to use for routing (-d in) or for queue binding (-d out)
  -c, --count=      The number of messages to send (10)
  -i, --interval=   The delay (in ms) between sending messages (10)
  -z, --size=       Message size in kB (1)
  -t, --stddev=     Standard deviation of message size (0)
  -r, --renew       Automatically resubscribe when the server cancels a subscription (used for mirrored queues) (false)
  -u, --user=       The user to connect as (guest)
  -w, --pass=       The user's password (guest)
  -H, --host=       The Rabbit host to connect to (localhost)
  -p, --port=       The Rabbit port to connect on (5672)
  -e, --entropy     Display message level entropy information (false)
  -V, --version     Print rmq version and exit

Help Options:
  -h, --help       Show this help message
```

Roadmap
-------

In no particular order:

* Integration with the RabbitMQ management API
* Rate limting
* Flow control

License
-------

The MIT License (MIT)

Copyright (c) [2014] [RelOps Ltd]

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
