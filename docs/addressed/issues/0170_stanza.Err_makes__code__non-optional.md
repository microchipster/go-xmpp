# Issue #170: stanza.Err makes 'code' non-optional

**Opened by** ivucica **at** 2020-12-14T18:47:43Z

## Body

Neither RFC3920 nor RFC6120 require the `code` attribute on `<error/>`, but the implementation of `stanza.Err` type makes it required, and silently so:

https://github.com/FluuxIO/go-xmpp/blob/947fcf0/stanza/error.go#L81-L83

That is: omitting code requires in the whole element silently not being marshalled.

The check for code==0 should merely result in the `code` attribute being omitted, not in the whole `<error/>` being omitted.

---
**Addressed note**

`stanza.Err.MarshalXML()` now omits the `code` attribute when `Code == 0` but still serializes the `<error/>` element and any other error payload fields. Regression coverage was added for code-less errors.

---
