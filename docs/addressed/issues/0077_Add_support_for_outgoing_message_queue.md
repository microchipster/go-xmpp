# Issue #77: Add support for outgoing message queue

**Opened by** mremond **at** 2019-06-26T14:01:23Z

## Body

We should allow sending messages for a while, until we reconnect. The send would fail on messages with an error when the queue is full.

We should also link that to stream management ack to keep outgoing message in that queue until they have been acked by the server, so that they can be replayed if they are lost.

## Comments

---
**Addressed note**

This is addressed in-tree by storing outbound stanzas in `Session.SMState.UnAckQueue` when stream management is enabled, and by replaying unacked stanzas from `SendMissingStz` after SM acknowledgements or stream resumption. `Send()` and `SendRaw()` both feed the queue, so outgoing messages are retained until the server acks them or the session is reconnected.

---
**remicorniere** at 2020-01-13T10:25:31Z
Maybe see #127 first
