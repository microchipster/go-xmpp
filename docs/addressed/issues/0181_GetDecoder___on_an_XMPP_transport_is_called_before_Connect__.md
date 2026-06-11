# Issue #181: GetDecoder() on an XMPP transport is called before Connect()

**Opened by** bodqhrohro **at** 2022-05-23T21:03:16Z

## Body

```
ERRO[0212] stream error: invalid-from

[ 3][t 0][1653306110.321923017][Client.cpp:291][&td_requests]   End to wait for updates, returning object 205 0x7fb9e43f8e50
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x90 pc=0x1173bba]

goroutine 573 [running]:
encoding/xml.(*Decoder).Token(0x0, 0x10, 0xc0002bfcb8, 0xb9bf79, 0x20)
/usr/lib/go-1.15/src/encoding/xml/xml.go:282 +0x3a
gosrc.io/xmpp/stanza.NextXmppToken(0x0, 0x7fba1c22b108, 0xc00038b8c0, 0x30, 0xc000366400)
/home/bodqhrohro/go/pkg/mod/dev.narayana.im/narayana/go-xmpp@v0.0.0-20211218155535-e55463fc9829/stanza/parser.go:89 +0x45
gosrc.io/xmpp/stanza.NextPacket(0x0, 0x0, 0x0, 0xc000464690, 0xc0002bffc0)
/home/bodqhrohro/go/pkg/mod/dev.narayana.im/narayana/go-xmpp@v0.0.0-20211218155535-e55463fc9829/stanza/parser.go:53 +0x45
gosrc.io/xmpp.(*Component).recv(0xc0001b0000)
/home/bodqhrohro/go/pkg/mod/dev.narayana.im/narayana/go-xmpp@v0.0.0-20211218155535-e55463fc9829/component.go:128 +0x93
created by gosrc.io/xmpp.(*Component).Resume
/home/bodqhrohro/go/pkg/mod/dev.narayana.im/narayana/go-xmpp@v0.0.0-20211218155535-e55463fc9829/component.go:104 +0xa0b
```

## Comments

---
**Addressed note**

This was addressed in-tree by making transports return a safe empty decoder before `Connect()` runs. `NewClientTransport` and `NewComponentTransport` now initialize a decoder up front, and `GetDecoder()` lazily fills in an empty decoder if needed so callers can safely probe parser state before a network stream exists.

---
**initpwn** at 2022-06-13T17:36:13Z
Please help to reproduce this error. 

---
**bodqhrohro** at 2022-06-13T22:24:27Z
It was literally an invalid `from` in an outgoing stanza sent from my code (AFAIR, ending in `@` with no server part).  

The problem though is that stream errors should be handled more gracefully than a segmentation fault, I suppose.
