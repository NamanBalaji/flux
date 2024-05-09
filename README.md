# Flux
Flux is a high-performance in memory event streaming platform offering scalable topic-based messaging, guaranteed message ordering, and at-least-once delivery.

![Build Status](https://github.com/NamanBalaji/flux/actions/workflows/ci.yml/badge.svg)
[![codecov](https://codecov.io/gh/NamanBalaji/flux/branch/main/graph/badge.svg?token=YOUR_TOKEN)](https://codecov.io/gh/<username>/<repository>)

## Features

- Topic based messaging
- In Memory storage
- Message ordering 
- At least once delivery 
- Message deduplication
- Periodic state cleanup

## Design

### Producers

- Producers send messages to the broker. 

#### Message Duplication and Ordering:

- Messages are structured like <message, topic, uuid>.
- Each message produced by producers is associated with a unique id decided by producer (uuid), which is used to prevent duplication: producers can retry sending a same message if ther's an error or request timeout.

### Consumers

#### Push Model and Subscription Semantics:

- Consumers can subscribe to some topics and get the messages related to the topics. We plan to use the push model for consumers, brokers push messages to consumers instead of consumers poll from brokers. The push model is better for our project because all the messages are in-memory, and the broker should deliver and purge the messages as soon as possible to save memory space.
- While subscribing consumers can send `readOld` flag which allows the consumer to read all the old messages that the broker stills has in memory before reading the new ones.

#### Consumer Registration and Message Delivery:

- Consumers register themselves with the broker (leader) when they start up and specify the topics they're interested in.
- New consumers can choose to receive only new messages or all the old messages that the broker has in it's record 
- We provide a semantics that the consumer receives the messages of a topic in a strictly identical order as in the brokers. So the broker will advance the offset only when a consumer acknowledges a message.

### Brokers

Brokers store and manage messages from producer and push messages to consumers. Brokers maintain a request channel to process the publishing requests in order.

#### In-Memory Storage:

- Brokers store messages in-memory. A thread safe queue data structure is used to store the messages.
- Brokers periodically clean up inactive subscribers and all the messages that have been ACKed.

#### Subscriber and Message Management:

- Consumers are set to inactive if they unsubscribe, or they fail ack when the broker pushes a message
- Subscribers are deleted from the memory if they have inactive status and they are last activity was recorded more than their ttl 
- Messages are only deleted if they are delivered to all the subscribed consumers and if their ttl is expired

## TODOs
- [ ] Add unit tests 
- [ ] Add benchmarks