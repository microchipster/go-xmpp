# Issue #147: Support Message Archive Management (XEP-0313)

**Opened by** remicorniere **at** 2020-01-28T09:58:02Z

## Body

https://xmpp.org/extensions/xep-0313.html

## Comments

---
**Addressed note**

This was addressed in-tree by adding MAM stanza types for query, fin, result, forwarded, and delay payloads, all registered for `urn:xmpp:mam:0/1/2` and covered by MAM marshal/unmarshal tests.
