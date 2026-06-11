package xmpp

import (
	"encoding/xml"
	"reflect"
	"testing"

	"gosrc.io/xmpp/stanza"
)

func TestStreamErrorParsesSeeOtherHostTarget(t *testing.T) {
	raw := `<error xmlns="http://etherx.jabber.org/streams"><see-other-host xmlns="urn:ietf:params:xml:ns:xmpp-stanzas">redirect.example:5223</see-other-host></error>`

	var packet stanza.StreamError
	if err := xml.Unmarshal([]byte(raw), &packet); err != nil {
		t.Fatalf("failed to unmarshal stream error: %v", err)
	}

	if got, want := packet.Error.Local, "see-other-host"; got != want {
		t.Fatalf("unexpected stream error name: got %q want %q", got, want)
	}
	if got, want := packet.SeeOtherHost, "redirect.example:5223"; got != want {
		t.Fatalf("unexpected see-other-host target: got %q want %q", got, want)
	}
}

func TestClientConnectionAddressesPrioritizeRedirect(t *testing.T) {
	client := &Client{
		config: &Config{
			TransportConfiguration: TransportConfiguration{
				Address: "original.example:5222",
			},
		},
		srvAddresses: []string{"srv1.example:5222", "srv2.example:5222"},
	}
	client.Session = &Session{
		SMState: SMState{preferredReconAddr: "redirect.example:5223"},
	}

	got := client.connectionAddresses()
	want := []string{"redirect.example:5223", "srv1.example:5222", "srv2.example:5222"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected connection address order: got %#v want %#v", got, want)
	}
}
