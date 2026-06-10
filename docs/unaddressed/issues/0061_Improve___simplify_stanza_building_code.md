# Issue #61: Improve / simplify stanza building code

**Opened by** mremond **at** 2019-06-20T13:38:34Z

## Body

Currently, building a stanza can lead to quite lengthy code, like the following:

```go
iqResp := xmpp.NewIQ(xmpp.Attrs{Type: "result", From: iq.To, To: iq.From, Id: iq.Id, Lang: "en"})
identity := xmpp.Identity{
	Name:     opts.Name,
	Category: opts.Category,
	Type:     opts.Type,
}
payload := xmpp.DiscoInfo{
	XMLName: xml.Name{
		Space: xmpp.NSDiscoInfo,
		Local: "query",
	},
	Identity: identity,
	Features: []xmpp.Feature{
		{Var: xmpp.NSDiscoInfo},
		{Var: xmpp.NSDiscoItems},
		{Var: "jabber:iq:version"},
		{Var: "urn:xmpp:delegation:1"},
	},
}
iqResp.Payload = &payload
```

I am exploring ways to make the code more compact and more guided through helpers (usable with code completin) with a code like this:

```go
b := stanza.NewBuilder().Lang("fr") // Create builder and set default language

iq := b.IQ(stanza.Attrs{Type: "get", To: "service.localhost", Id: "disco-get-1"})

payload := b.DiscoInfo()
identity := b.Identity("Test Component", "gateway", "service")
payload.SetFeatures(stanza.NSDiscoInfo, stanza.NSDiscoItems, "jabber:iq:version", "urn:xmpp:delegation:1").
	SetIdentities(identity)

iq.Payload = payload
```

What do you think?

## Comments

---
**mremond** at 2019-06-20T13:41:34Z
@genofire Any opinion on this?
Do you like the prototype API to create packets better ?

---
**genofire** at 2019-06-20T13:53:46Z
unnecessary complicated
for me the source code is easier to read, if it is formated like before.

---
**mremond** at 2019-06-20T14:14:20Z
Thanks for feedback.
I committed my exploration in that branch: https://github.com/FluuxIO/go-xmpp/tree/builder-exploration/pkg/stanza to give you a better idea.

I feel that at some point the code may be difficult to navigate, with the pure parsing / packet formatting part, being mixed with client, component workflow and connection manager.

Separating the stanza building, marshalling and unmarshalling maybe help structure the code better as it grows (and I did not add test modules for all the structure generation, yet).

I think it should still be possible to use the full expending stanza version as well.

Anyway, still thinking about it, as you find this complicated. Maybe I will end up doing something in between (move to a separate package but remove the builder interface).

I am also afraid that godoc will become unusable if we do no split the XML marshalling / unmarshalling part: https://godoc.org/gosrc.io/xmpp

---
**mremond** at 2019-06-27T10:04:47Z
Using pointers, I managed to have a prototype of approach that can further reduce building an IQ with payload, as follows:

```go
iq := stanza.NewIQ(stanza.Attrs{Type: "get", To: "service.localhost", Id: "disco-get-1"})
disco := iq.DiscoInfo()
disco.AddIdentity("Test Component", "gateway", "service")
disco.AddFeatures(stanza.NSDiscoInfo, stanza.NSDiscoItems, "jabber:iq:version", "urn:xmpp:delegation:1")
```

---
**mremond** at 2019-06-27T13:20:35Z
@genofire Several things to note for upgrade:
1. There is now a stanza package to host parsing and stanza structure. It should help make the project structure more obvious.
2. On DiscoInfo package, you can have several identities. For future reference, see example 2 in XEP-0030: https://xmpp.org/extensions/xep-0030.html#example-2

Other than that, you are not force to use shortcuts and can keep on relying on building the stanza with structs, as before.
