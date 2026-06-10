# Issue #171: Golang XML vulnerabilities

**Opened by** mdosch **at** 2020-12-15T09:09:02Z

## Body

FYI

https://github.com/mattermost/xml-roundtrip-validator/blob/master/advisories/unstable-attributes.md
https://github.com/mattermost/xml-roundtrip-validator/blob/master/advisories/unstable-directives.md
https://github.com/mattermost/xml-roundtrip-validator/blob/master/advisories/unstable-elements.md

## Comments

---
**licaon-kter** at 2020-12-15T12:07:22Z
Ref: https://mattermost.com/blog/coordinated-disclosure-go-xml-vulnerabilities/

---
**Neustradamus** at 2020-12-15T15:28:08Z
@mdosch, @licaon-kter: Thanks for this information :)

---
**prefiks** at 2020-12-15T15:55:52Z
Not sure if this is really problem for this library, i think what this vulnerability is that if you have element like `<test:a xmlns:test="abc" xmlns:test2="abc"/>` after parsing and serializing it back you could get `<test2:a xmlns:test="abc" xmlns:test2="abc"/>`, but i don't think this will be problem for us - we don't require original namespaces, and i don't even think we serialize back values that we parsed previously.
