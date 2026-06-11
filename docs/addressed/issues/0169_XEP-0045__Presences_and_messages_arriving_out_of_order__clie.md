# Issue #169: XEP-0045: Presences and messages arriving out of order; clients can interpret room join as complete before all `<presence/>`s landed

**Opened by** ivucica **at** 2020-09-28T19:26:18Z

## Body

Actors:

- `example.org`: host domain
- `freenode.example.org`: MUC domain (proxying traffic to IRC)
- `abc@example.org`: bot joining the MUC room
- `#abcd@freenode.example.org`: MUC room being joined

This is the traffic received, anonymized and shortened, mildly formatted (note, all of this is listed under one `RECV:`):

```
<presence to='abc@example.org/abcd-go' from='#abcd@freenode.example.org/_joe'>
  <x xmlns='http://jabber.org/protocol/muc#user'>
    <item jid='_joe@freenode.example.org' affiliation='none' role='participant'/>
  </x>
</presence>

<presence to='abc@example.org/abcd-go' from='#abcd@freenode.example.org/abc'>
  <x xmlns='http://jabber.org/protocol/muc#user'>
    <status code='110'/>
    <item jid='abc@freenode.example.org/75440f1696@machine.example.org' affiliation='none' role='participant'/>
  </x>
</presence>

<message type='groupchat' to='abc@example.org/abcd-go' from='#abcd@freenode.example.org/PersonSettingSubject'>
  <subject>this is some subject</subject>
</message>
```

However, the very first packet landing into `HandlePacket` is the `<message/>`. Per XEP-0045, a `<message/>` with no `<body/>` _[empty not being a good replacement]_ and containing a mandatory `<subject/>` _[which may be empty; see also https://github.com/FluuxIO/go-xmpp/issues/167]_ means that the room join is complete.

Yet, none of the `<presence/>`s have been seen at this point.

Inability to distinguish between empty and not-present `body` and `subject`   https://github.com/FluuxIO/go-xmpp/issues/167 and delivery of `<message/>` and `<presence/>` out of order means it's not possible to know which users were already present in the room and which entered later, despite XEP-0045 allowing for this scenario.

This is on ac5b066815b21708dacd62a169508d0aa408cee6 from May 7 2020, currently the latest commit.

---
**Addressed note**

The client receive loop now routes packets inline instead of spawning a goroutine, so sequential presence/message stanzas are observed in arrival order. Combined with the pointer-backed message fields from `#0167`, MUC join flows can distinguish empty subject/body stanzas without reordering them away from the corresponding presence updates.

---
