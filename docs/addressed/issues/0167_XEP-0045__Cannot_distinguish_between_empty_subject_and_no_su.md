# Issue #167: XEP-0045: Cannot distinguish between empty subject and no subject

**Opened by** ivucica **at** 2020-08-25T20:27:42Z

## Body

Hi,

per https://xmpp.org/extensions/xep-0045.html#enter-subject an empty subject is permitted and necessary to complete the join process in case of no subject being specified.

> If there is no subject set, the room MUST return an empty &lt;subject/&gt; element.

However, in the `Message` stanza, `Subject` field is a plain string (not a pointer), and the `xml:` field tag contains `,omitempty`.

This makes it night impossible to spot an empty-subject message (or, for that matter, an empty-but-present body message).

(It might be doable by using a custom `MsgExtension` and doing a `Get()` over it, but this is inconvenient, and if possible at all, I cannot quickly think of a way to do it, given that a global type registry is involved.)

An improvement would be to simply replace the `string` with a `*string`, and evaluating whether to use `omitempty`.

## Comments

---
**ivucica** at 2020-08-25T20:35:31Z
Additionally, empty vs non-present `<body/>` is relevant for XEP-0045:

> Note: In accordance with the core definition of XML stanzas, any message can contain a `<subject/>` element; only a message that contains a `<subject/>` but no `<body/>` element shall be considered a subject change for MUC purposes.

---
**Addressed note**

`stanza.Message` now keeps `Subject` and `Body` as pointer-backed fields, and the custom XML unmarshal path preserves empty-but-present `<subject/>` and `<body/>` elements while leaving absent ones nil. Regression tests cover both the empty and missing cases.

---
