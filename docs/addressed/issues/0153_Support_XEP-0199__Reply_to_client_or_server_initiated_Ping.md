# Issue #153: Support XEP-0199: Reply to client or server initiated Ping

**Opened by** remicorniere **at** 2020-02-28T14:09:11Z

## Body

Support for https://xmpp.org/extensions/xep-0199.html
Suggested by @vduduh 

## Comments

---
**Addressed note**

This was addressed in-tree by adding `stanza.Ping`, registering it as `urn:xmpp:ping`, and teaching the router to auto-reply to incoming IQ-get ping requests with an empty IQ result.
