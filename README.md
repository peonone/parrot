# Parrot 

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