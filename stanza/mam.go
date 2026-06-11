package stanza

import (
	"encoding/xml"
	"strings"
)

const (
	NSMAM0  = "urn:xmpp:mam:0"
	NSMAM1  = "urn:xmpp:mam:1"
	NSMAM2  = "urn:xmpp:mam:2"
	NSMAM   = NSMAM2
	NSXData = "jabber:x:data"
	NSDelay = "urn:xmpp:delay"
)

var supportedMAMNamespaces = []string{NSMAM2, NSMAM1, NSMAM0}

type MAMQuery struct {
	XMLName xml.Name   `xml:"query"`
	QueryID string     `xml:"queryid,attr,omitempty"`
	Form    Form       `xml:"x"`
	Set     *ResultSet `xml:"set,omitempty"`
}

func (q *MAMQuery) Namespace() string {
	return q.XMLName.Space
}

func (q *MAMQuery) GetSet() *ResultSet {
	return q.Set
}

type MAMFin struct {
	XMLName  xml.Name   `xml:"fin"`
	Complete bool       `xml:"complete,attr,omitempty"`
	Set      *ResultSet `xml:"set,omitempty"`
}

func (f *MAMFin) Namespace() string {
	return f.XMLName.Space
}

func (f *MAMFin) GetSet() *ResultSet {
	return f.Set
}

type MAMResult struct {
	MsgExtension
	XMLName   xml.Name      `xml:"result"`
	QueryID   string        `xml:"queryid,attr,omitempty"`
	ID        string        `xml:"id,attr,omitempty"`
	Forwarded *MAMForwarded `xml:"forwarded"`
}

type MAMForwarded struct {
	XMLName xml.Name  `xml:"urn:xmpp:forward:0 forwarded"`
	Delay   *MAMDelay `xml:"delay"`
	Message *Message  `xml:"message"`
}

type MAMDelay struct {
	XMLName xml.Name `xml:"urn:xmpp:delay delay"`
	Stamp   string   `xml:"stamp,attr,omitempty"`
}

func init() {
	for _, ns := range supportedMAMNamespaces {
		TypeRegistry.MapExtension(PKTIQ, xml.Name{Space: ns, Local: "query"}, MAMQuery{})
		TypeRegistry.MapExtension(PKTIQ, xml.Name{Space: ns, Local: "fin"}, MAMFin{})
		TypeRegistry.MapExtension(PKTMessage, xml.Name{Space: ns, Local: "result"}, MAMResult{})
	}
}

func (m *MAMResult) hasQueryID(id string) bool {
	return strings.TrimSpace(m.QueryID) == strings.TrimSpace(id)
}
