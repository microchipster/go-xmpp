# Issue #84: Add support for see-other-host server directive

**Opened by** mremond **at** 2019-06-29T14:01:23Z

## Body

The see-other-host stream level directive is used by a server to instruct a client to reconnect on another host. It is for example used when a node in a cluster is shutting down for maintenance operations.

Reference: https://tools.ietf.org/html/rfc6120#section-4.9.3.19

---
**Addressed note**

The XMPP stream-error parser now preserves the `see-other-host` payload, and the client/component reconnect flow uses that target before SRV/original fallback on the next resume attempt. `StreamManager` now avoids double-resume behavior after stream errors, so the redirect path is applied once and cleanly. Regression tests cover the payload parsing and reconnect address ordering.

---
