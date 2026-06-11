# Issue #178: InitialPresence should not be a constant

**Opened by** ivucica **at** 2021-11-09T22:43:30Z

## Body

I'd like to include XEP-0086 caps in the `<presence/>` sent. (Some users may want to also set a different `<show/>`, `<status/>` or `<priority/>`).

Right now, it's a constant that cannot be adjusted, so I'll send one immediately after -- however, it should be adjustable before the connection starts.

---
**Addressed note**

`Config` now exposes `InitialPresence`, and `postSessionSetup()` sends that stanza after the connection is established, falling back to the default `<presence/>` when the field is empty. Tests cover a custom initial presence with `show`, `status`, and `priority` elements.

---
