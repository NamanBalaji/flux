# flux
Flux is a high-performance, fault-tolerant distributed event streaming platform, featuring scalable topic-based messaging, and robust replication to ensure data consistency and availability.

## Todo
-[ ] add logging
-[ ] state dump
-[ ] subscriber cleanup improve logic

## Design

### Producers

#### Leader Election and Message Routing:

- Producers send messages to a leader, the leader is determined through the Raft consensus algorithm among brokers. Each broker group runs its own Raft consensus to elect a leader within the group.
- Producers can send a message to any broker. If the broker is not the leader, it redirects the producer to the current leader of the group.

#### Message Duplication and Ordering:

- Messages are structured like <message, topic, uuid>.
- Each message produced by producers is associated with a unique id decided by producer (uuid), which is used to prevent duplication: producers can retry sending a same message if they do not hear a response for some while and the receivers remember each message received with the uuid, and if duplicate messages are received, simple acknowledge it without appending it to its log.
- The max_in_flight configuration remains, allowing producers to specify how many messages can be sent asynchronously.


### Consumers

#### Push Model and Subscription Semantics:

- Consumers can subscribe to some topics and get the messages related to the topics. We plan to use the push model for consumers, namely, brokers push messages to consumers instead of consumers poll from brokers. The push model is better for our project because all the messages are in-memory, and the broker should deliver and purge the messages as soon as possible to save memory space.
- We plan to first develop a one-to-all semantics for consumers: a message will be sent to all consumers (e.g. a chat room application), we can add a one-to-one semantics (a message is sent to any and only one consumer who subscribed to the topic) later if time allows.

#### Consumer Registration and Message Delivery:

- Consumers register themselves with the broker (leader) when they start up and specify the topics they're interested in.
- New consumers can receive only new messages or old messages and then new messages produced after their registration.
- We provide a semantics that the consumer receives the messages of a topic in a strictly identical order as in the brokers. So the broker will advance the offset only when a consumer acknowledges a message.

### Brokers and Raft

Brokers store and manage messages from producer, replicate messages to other broker servers and push messages to consumers.

#### In-Memory Storage and Replication:

- Brokers store messages in-memory. The Raft protocol ensures that all messages are replicated across the broker group before being acknowledged to the producer.
- Each broker group runs an instance of the Raft consensus algorithm, with one leader and multiple followers at any given time.

#### Raft for Leader Election and Failover:

- Brokers within a group use Raft to elect a leader and to manage leader failover transparently.
- Leaders are responsible for replicating the log (messages) to followers and for pushing messages to consumers.

#### Consumer Offsets and Message Purging:

- The leader broker manages consumer offsets and periodically commits this information to the replicated log to ensure it's consistent across followers.
- Messages are purged from the broker storage only when all consumers for a topic have acknowledged receipt, ensuring no loss of messages on consumer failure.

#### Raft Cluster Information

- Broker Groups as Raft Clusters: Each broker group functions as a separate Raft cluster, managing its own leader election and log replication.
- Replication and Consensus: Raft ensures that all changes (message arrivals, consumer offsets) are replicated across all brokers in a group before being committed.
