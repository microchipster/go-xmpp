# Issue #106: SCRAM-SHA-1(-PLUS) + SCRAM-SHA-256(-PLUS) supports

**Opened by** Neustradamus **at** 2019-09-05T08:20:44Z

## Body

"When using the SASL SCRAM mechanism, the SCRAM-SHA-256-PLUS variant SHOULD be preferred over the SCRAM-SHA-256 variant, and SHA-256 variants [RFC7677] SHOULD be preferred over SHA-1 variants [RFC5802]".

SCRAM-SHA-1(-PLUS):
- https://tools.ietf.org/html/rfc5802
- https://tools.ietf.org/html/rfc6120

SCRAM-SHA-256(-PLUS):
- https://tools.ietf.org/html/rfc7677 since 2015-11-02
- https://tools.ietf.org/html/rfc8600 since 2019-06-21: https://mailarchive.ietf.org/arch/msg/ietf-announce/suJMmeMhuAOmGn_PJYgX5Vm8lNA

I add SCRAM-SHA-512(-PLUS): https://xmpp.org/extensions/inbox/hash-recommendations.html

Linked to:
- https://github.com/scram-xmpp/info/issues/1

## Comments

---
**genofire** at 2019-09-06T12:48:40Z
@mremond could we use https://godoc.org/mellium.im/sasl
or should we implemented it on the own like https://github.com/ortuman/jackal/blob/master/auth/scram.go

i like the state machine of https://godoc.org/mellium.im/sasl with his `more, resp, err` as return value

---
**SamWhited** at 2019-09-06T14:41:01Z
> i like the state machine of https://godoc.org/mellium.im/sasl with his more,
> resp, err as return value

Thanks; just FYI there is at least one major API overhaul that may be done at
some point to `mellium.im/sasl` to make credentials more flexible. Take into
consideration that I don't yet provide any API stability guarantees when using
it, and it may require a bit of work to upgrade later (that being said, that's
probably better than an unknown from some of the other projects, so maybe see
if you can convince them to tell people if the API is stable or not before
deciding what to move forward with).

If you do end up using it, I'd love to get your feedback on everything and how
it works for you.

> I add SCRAM-SHA-512(-PLUS): https://xmpp.org/extensions/inbox/hash-recommendations.html

Note that SCRAM-SHA-512 is not an actual SCRAM mechanism defined by the IETF. Making one up will at best lead to interop problems, and at worst could cause security issues (this isn't likely if you're just changing the hash in the HMAC, but still, don't make up security mechanisms just to have a bigger number in a hash somewhere).
