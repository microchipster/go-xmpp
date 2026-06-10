# PR #185: stanza: fix dropped error

**Opened by** alrs **at** 2025-07-03T19:38:52Z

## Body

This fixes a dropped `err` variable in the `stanza` package.

## Issue comments

---
**Addressed note**

This was addressed in-tree by making `stanza.Forwarded.UnmarshalXML` return the `decodeClient` error instead of silently ignoring it when a forwarded stanza cannot be parsed.
