package stanza

import "encoding/xml"

// Ping implements XEP-0199 ping.
type Ping struct {
	XMLName xml.Name `xml:"urn:xmpp:ping ping"`
}

func (Ping) Name() string {
	return "ping"
}

func (Ping) Namespace() string {
	return NSPing
}

func (Ping) GetSet() *ResultSet {
	return nil
}

const NSPing = "urn:xmpp:ping"

func init() {
	TypeRegistry.MapExtension(PKTIQ, xml.Name{Space: NSPing, Local: "ping"}, Ping{})
}
