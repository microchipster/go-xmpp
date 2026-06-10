# PR #190: Reconnect fixes

**Opened by** mwild1 **at** 2025-10-09T09:27:14Z

## Body

This PR includes multiple small fixes go-xmpp to successfully reconnect when the server sends a `</stream:stream>`.

## Issue comments

---
**Neustradamus** at 2026-02-10T21:42:27Z
@mremond: Have you seen this PR?

---

**Addressed note**

This reconnect fix was addressed in-tree by buffering the transport stream-close notification, making client-initiated disconnects mark the session `StateDisconnected`, and routing server-initiated `</stream:stream>` handling through the same disconnect path so stream management can resume cleanly.
