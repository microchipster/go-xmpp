# Issue #85: Session establishment should timeout if server does not send any expected response

**Opened by** mremond **at** 2019-06-29T15:37:18Z

## Body

---
**Addressed note**

Session establishment now uses transport deadlines driven by `ConnectTimeout`. The client applies a deadline before stream negotiation and SASL/bind/session setup, and the component applies the same during handshake, so missing server replies fail with a timeout instead of hanging indefinitely.

---


