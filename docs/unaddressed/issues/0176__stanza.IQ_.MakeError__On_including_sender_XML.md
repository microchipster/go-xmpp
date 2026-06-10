# Issue #176: (stanza.IQ).MakeError: On including sender XML

**Opened by** ivucica **at** 2021-04-07T18:37:13Z

## Body

Per [RFC 6120 8.3.2.](https://tools.ietf.org/html/rfc6120#section-8.3.2), including the 'sender XML' is optional, however it seems deceptively easy to do with Fluux XMPP, given that `iq` of type `get` and `set` are defined to have only one[^1] child element per [RFC 6120 8.2.3](https://tools.ietf.org/html/rfc6120#section-8.2.3)

As including sender XML in the response is optional, it's present in various XEP examples and *may* help debugging, so I explored including it -- even if I don't require this myself and I don't think any clients require responses to include it.

However, trivial assignment such as `iqErr.Error = iq.Error` would result in including a fully-deserialized-then-reserialized XML making this vulnerable to

1. bugs in `XMLName` on the backing struct for the payload (that is: it doesn't match the type registry values)
1. other mismatches in the struct resulting in non-equivalent XML being generated on re-serialization
1. more generally: the unstable-{attributes,directives,elements} flaw related to namespaces: https://github.com/FluuxIO/go-xmpp/issues/171.

Therefore: `stanza.IQ` could store the original bytes that the sender transmitted (taking care that overly-large stanzas don't DoS the service), and `(stanza.IQ).MakeError` could copy the bytes verbatim. 

This itself may have the problem of inheriting the wrong `xmlns` depending on how the transmitting client defined the namespace prefixes[^3]. Of course this may be overthinking it: I don't believe this data should be interpreted by any reasonable client, so maybe just including a copy captured during deserialization would be enough?

Does it make sense to make `MakeError` do the 'correct' thing, so devs are not inclined to assigning `iq.Payload` to `iqErr.Payload` resulting in incorrect 'sender XML' being sent over the wire? It may be too much effort, so perhaps just including a docstring explaining why the payload is not included would be sufficient?

---

[^1]: This gets more complicated with error-responses from `message` and `presence`, but I was mainly thinking of `iq`.
[^2]: It may even require addition of the payload's `xmlns` in case namespaces don't match.
[^3]: Can this be resolved by forcibly including `xmlns` from the original `Payload` if there's a mismatch from the `iq` stanza's namespace?

