# Issue #95: Improve DNS SRV resolution by adding fallback if first server does not work

**Opened by** mremond **at** 2019-07-27T16:12:35Z

## Body

This is a follow up of #94 

---
**Addressed note**

The client already resolves all SRV targets in priority order and retries through the full address list in `resolveSRVAddresses()` and `Client.connect()`, so the fallback behavior described here is already covered by the connection logic.

---
