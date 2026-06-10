# Issue #84: Add support for see-other-host server directive

**Opened by** mremond **at** 2019-06-29T14:01:23Z

## Body

The see-other-host stream level directive is used by a server to instruct a client to reconnect on another host. It is for example used when a node in a cluster is shutting down for maintenance operations.

Reference: https://tools.ietf.org/html/rfc6120#section-4.9.3.19
