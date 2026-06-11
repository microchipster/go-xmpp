# Issue #173: "Echo" example from README should following RFC6121 

**Opened by** adefirmanf **at** 2020-12-17T09:49:48Z

## Body

Based on RFC6121, attributes for sending message from README and example doesn't following the documentation. It produce the message doesn't receive to client. 

I would suggest to PR the example for update README and example library. Adding "chat" as type will resolve this issue. 

`stanza.Message{Attrs: stanza.Attrs{To: msg.From, Type: "chat"}, Body: msg.Body}`

Reference : https://xmpp.org/rfcs/rfc6121.html#message-syntax

---
**Addressed note**

The README echo example and `_examples/xmpp_echo/xmpp_echo.go` now send replies with `Type: stanza.MessageTypeChat`, which matches RFC6121 message semantics and restores delivery to regular chat clients.

---
