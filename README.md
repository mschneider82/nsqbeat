# Nsqbeat

[![Build Status](https://travis-ci.org/mschneider82/nsqbeat.svg?branch=master)](https://travis-ci.org/mschneider82/nsqbeat)

Nsqbeat is an elastic [Beat](https://www.elastic.co/products/beats) that reads
events from one [NSQ](https://nsq.io) topic and forwards them to
[Logstash](https://www.elastic.co/products/logstash) (or any other configured output like elasticsearch).

The NSQ consumer implements an at-least-once behaviour which means that
messages may be forwarded to the configured output more than once.

## Getting Started with Nsqbeat

### Requirements

* [Golang](https://golang.org/dl/) 1.7

### Building

```sh
# Make sure $GOPATH is set
go get github.com/mschneider82/nsqbeat
cd $GOPATH/src/github.com/mschneider82/nsqbeat
make
```

### Running

To run Nsqbeat with debugging output enabled, run:

```sh
./nsqbeat -c nsqbeat.yml -e -d "*"
```

### Installing

You can use the [precompiled packages](https://github.com/mschneider82/nsqbeat/releases)

### Configuring

An example configuration can be found in the file `nsqbeat.yml`. The following
parameters are specific to Nsqbeat:

```yaml
nsqbeat:
    # A list of NSQ Lookup Daemons to connect to
    lookupdhttpaddrs: ["127.0.0.1:4161"]
    # a Topic to sucscribe to
    topic: "sometopic"
    # The channel name to join
    channel: "testchan"
    # How many in Flights
    maxinflight: 200
    # If data in the topic is Json then use the decoder, if not set to something else like plain
    codec: "json"
    # use Golang time format layout to define if @timestamp exists and has a different format
    timelayout: "2006-01-02T15:04:05.000Z"

```

### Testing

To test Nsqbeat, run the following command:

```sh
make testsuite
```

alternatively:

```sh
make unit-tests
make system-tests
make integration-tests
make coverage-report
```

The test coverage is reported in the folder `./build/coverage/`
