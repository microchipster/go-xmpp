# Issue #34: Client should keep track of XMPP server features

**Opened by** mremond **at** 2019-06-05T09:17:23Z

## Body


## Comments

---
**Addressed note**

The client now preserves pre-auth stream features separately from post-auth features, and the current implementation exposes those preserved server features through `Client.ServerFeatures()` and `Session.ServerFeatures`.

