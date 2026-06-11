# Issue #64: Add ability to set the preferred lang at the stream level

**Opened by** mremond **at** 2019-06-20T16:07:07Z

## Body


---
**Addressed note**

The client now threads `Config.Lang` into stream-open framing for native TCP, websocket, and cert-checker connections by emitting `xml:lang` when the language is configured. `TransportConfiguration` carries the stream language, `stanza.Stream` reflects it with an `xml:lang` attribute, and regression tests cover both the copied config value and the generated stream open strings.

---
