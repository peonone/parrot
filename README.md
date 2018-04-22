# Parrot [![Build Status](https://travis-ci.org/peonone/parrot.svg?branch=master)](https://travis-ci.org/peonone/parrot) [![Go Report Card](https://goreportcard.com/badge/peonone/parrot)](https://goreportcard.com/report/github.com/peonone/parrot)

parrot is a microservice based chat server for practice purpose, it's powered by [github.com/micro/go-micro](https://github.com/micro/go-micro).

## architecture
```
                       +------------+             +-----------------+
                HTTP   |            |             |                 |
              +------->+Auth Web API+------------>+ Auth Service    |
              |        |            |       +----->                 |
              |        +------------+       |     +-----------------+
              |                             |
+-------------+                             |
|             |        +--------------------+     +----------------+
|   browser   |        |                    |     |                |
|             |   WS   |Chat Websocket server+---->Chat Service    |
+-------------+------->+                    |     |                |
                       +--------------------+     +----------------+
                                           ^              |
                                           |              \/
                                       +---------------------------------------+
                                       |                                       |
                                       | RabbitMQ (Topic Exchange)             |
                                       | For push message back to ws client    |
                                       +---------------------------------------+
```