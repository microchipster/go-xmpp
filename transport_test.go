package xmpp

import (
	"strings"
	"testing"

	"gosrc.io/xmpp/stanza"
)

func TestNewClientTransportGetDecoderBeforeConnect(t *testing.T) {
	tests := []struct {
		name string
		addr string
	}{
		{name: "tcp", addr: "example.org"},
		{name: "websocket", addr: "ws://example.org/xmpp-websocket"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			transport := NewClientTransport(TransportConfiguration{Address: tc.addr})
			decoder := transport.GetDecoder()
			if decoder == nil {
				t.Fatal("expected a decoder before connect")
			}

			if _, err := stanza.NextPacket(decoder); err == nil || err.Error() != "connection closed" {
				t.Fatalf("expected empty pre-connect decoder to return connection closed, got %v", err)
			}
		})
	}
}

func TestNewComponentTransportGetDecoderBeforeConnect(t *testing.T) {
	transport, err := NewComponentTransport(TransportConfiguration{Address: "example.org"})
	if err != nil {
		t.Fatalf("NewComponentTransport returned error: %v", err)
	}

	decoder := transport.GetDecoder()
	if decoder == nil {
		t.Fatal("expected a decoder before connect")
	}

	if _, err := stanza.NextPacket(decoder); err == nil || err.Error() != "connection closed" {
		t.Fatalf("expected empty pre-connect decoder to return connection closed, got %v", err)
	}
}

func TestNewClientCopiesLangIntoTransportConfiguration(t *testing.T) {
	config := Config{
		TransportConfiguration: TransportConfiguration{
			Address: "example.org:5222",
		},
		Jid:        "test@example.org",
		Credential: Password("secret"),
		Lang:       "en",
	}

	client, err := NewClient(&config, NewRouter(), clientDefaultErrorHandler)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	if got, want := client.config.TransportConfiguration.Lang, "en"; got != want {
		t.Fatalf("transport language = %q, want %q", got, want)
	}
}

func TestStreamOpenIncludesLanguageWhenConfigured(t *testing.T) {
	if got := clientStreamOpen("example.org", "en"); !strings.Contains(got, "xml:lang='en'") {
		t.Fatalf("client stream open missing language: %s", got)
	}
	if got := componentStreamOpen("example.org", "en"); !strings.Contains(got, "xml:lang='en'") {
		t.Fatalf("component stream open missing language: %s", got)
	}
	if got := websocketStreamOpen("example.org", "en"); !strings.Contains(got, "xml:lang='en'") {
		t.Fatalf("websocket stream open missing language: %s", got)
	}
	if got := clientStreamOpen("example.org", ""); strings.Contains(got, "xml:lang") {
		t.Fatalf("client stream open should omit language when empty: %s", got)
	}
}
