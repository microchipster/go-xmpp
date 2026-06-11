package stanza

import (
	"encoding/xml"
	"testing"
)

func TestMAMQueryAndFinExtensions(t *testing.T) {
	const queryXML = `<iq type='set' id='1'><query xmlns='urn:xmpp:mam:2' queryid='q1'><x xmlns='jabber:x:data' type='submit'><field var='FORM_TYPE' type='hidden'><value>urn:xmpp:mam:2</value></field></x><set xmlns='http://jabber.org/protocol/rsm'><max>10</max></set></query></iq>`

	var iq IQ
	if err := xml.Unmarshal([]byte(queryXML), &iq); err != nil {
		t.Fatalf("failed to unmarshal MAM query: %v", err)
	}
	query, ok := iq.Payload.(*MAMQuery)
	if !ok {
		t.Fatalf("expected MAMQuery payload, got %#v", iq.Payload)
	}
	if query.QueryID != "q1" {
		t.Fatalf("unexpected query id: %q", query.QueryID)
	}
	if query.Namespace() != NSMAM2 {
		t.Fatalf("unexpected query namespace: %q", query.Namespace())
	}
	if query.Set == nil || query.Set.Max == nil || *query.Set.Max != 10 {
		t.Fatalf("unexpected query result set: %#v", query.Set)
	}

	const finXML = `<iq type='result' id='1'><fin xmlns='urn:xmpp:mam:2' complete='true'><set xmlns='http://jabber.org/protocol/rsm'><count>4</count></set></fin></iq>`
	iq = IQ{}
	if err := xml.Unmarshal([]byte(finXML), &iq); err != nil {
		t.Fatalf("failed to unmarshal MAM fin: %v", err)
	}
	fin, ok := iq.Payload.(*MAMFin)
	if !ok {
		t.Fatalf("expected MAMFin payload, got %#v", iq.Payload)
	}
	if !fin.Complete {
		t.Fatal("expected fin to be complete")
	}
	if fin.Set == nil || fin.Set.Count == nil || *fin.Set.Count != 4 {
		t.Fatalf("unexpected fin result set: %#v", fin.Set)
	}
}

func TestMAMResultMessageExtension(t *testing.T) {
	const messageXML = `<message from='archive@example.com' to='user@example.com'><result xmlns='urn:xmpp:mam:2' queryid='q2' id='msg-1'><forwarded xmlns='urn:xmpp:forward:0'><delay xmlns='urn:xmpp:delay' stamp='2026-06-11T01:23:45Z'/><message xmlns='jabber:client' from='alice@example.com' to='user@example.com'><body>Hello</body></message></forwarded></result></message>`

	var msg Message
	if err := xml.Unmarshal([]byte(messageXML), &msg); err != nil {
		t.Fatalf("failed to unmarshal MAM result message: %v", err)
	}
	if len(msg.Extensions) != 1 {
		t.Fatalf("expected 1 extension, got %d", len(msg.Extensions))
	}
	res, ok := msg.Extensions[0].(*MAMResult)
	if !ok {
		t.Fatalf("expected MAMResult extension, got %#v", msg.Extensions[0])
	}
	if !res.hasQueryID("q2") || res.ID != "msg-1" {
		t.Fatalf("unexpected mam result identifiers: %#v", res)
	}
	if res.Forwarded == nil || res.Forwarded.Delay == nil || res.Forwarded.Delay.Stamp != "2026-06-11T01:23:45Z" {
		t.Fatalf("unexpected forwarded delay: %#v", res.Forwarded)
	}
	if res.Forwarded.Message == nil || res.Forwarded.Message.Body != "Hello" {
		t.Fatalf("unexpected forwarded message: %#v", res.Forwarded.Message)
	}
}
