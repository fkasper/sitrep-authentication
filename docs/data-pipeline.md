# Data pipeline of incoming data

```
Raw TY_HBS (from client) -> Websockets Connection ->
Handlebars parser (node) -> Kafka -> Reactive Streams Optimizer
          / Mongodb (html)
-(FANOUT)-  Mongodb (js)
          \ Mongodb (css)
- PURGE CACHE -> notify user (websockets)
```

## Raw TY_HBS Language

## Websocket Connection

## Handlebars parser

## Kafka Queue

## Reactive Streams Optimizer

## Fanout to Mongodb

## Purge CACHE

## Notify the user about the applied change
