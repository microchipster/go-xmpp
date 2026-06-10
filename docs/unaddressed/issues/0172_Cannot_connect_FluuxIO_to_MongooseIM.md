# Issue #172: Cannot connect FluuxIO to MongooseIM

**Opened by** theanhoo **at** 2020-12-16T11:25:49Z

## Body

I can't seem to connect FluuxIO to a MongooseIM XMPP server. I kept getting "NextStart XML syntax error on line 1: unexpected EOF"

Setting "Insecure" to true or false makes no difference.

Any advice?

Many thanks in advance.

```
SEND:
<?xml version='1.0'?><stream:stream to='example.com' xmlns='jabber:client' xmlns:stream='http://etherx.jabber.org/streams' version='1.0'>

RECV:
<?xml version='1.0'?><stream:stream xmlns='jabber:client' xmlns:stream='http://etherx.jabber.org/streams' id='E7EE923D9E43805D' from='example.com' version='1.0' xml:lang='en'>

RECV:
<stream:features><starttls xmlns='urn:ietf:params:xml:ns:xmpp-tls'/><mechanisms xmlns='urn:ietf:params:xml:ns:xmpp-sasl'><mechanism>SCRAM-SHA-512</mechanism><mechanism>SCRAM-SHA-384</mechanism><mechanism>SCRAM-SHA-256</mechanism><mechanism>SCRAM-SHA-224</mechanism><mechanism>SCRAM-SHA-1</mechanism><mechanism>PLAIN</mechanism></mechanisms><c xmlns='http://jabber.org/protocol/caps' hash='sha-1' node='https://www.erlang-solutions.com/products/mongooseim.html' ver='Ouw/7YxQV/T0hISjVQrLc+HGtQ8='/><register xmlns='http://jabber.org/features/iq-register'/><sm xmlns='urn:xmpp:sm:3'/></stream:features>

SEND:
<starttls xmlns='urn:ietf:params:xml:ns:xmpp-tls'/>

RECV:
<proceed xmlns='urn:ietf:params:xml:ns:xmpp-tls'/>

SEND:
</stream:stream>

NextStart XML syntax error on line 1: unexpected EOF
```

