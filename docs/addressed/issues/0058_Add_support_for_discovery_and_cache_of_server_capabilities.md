# Issue #58: Add support for discovery and cache of server capabilities

**Opened by** mremond **at** 2019-06-19T09:46:59Z

## Body

This is also related to cache of server features (see #34)

## Comments

---
**Addressed note**

The client now caches `disco#info` results in memory, keyed by the advertised caps hash when available and falling back to the server domain. Repeated capability lookups reuse the cached discovery result instead of querying the server again.
