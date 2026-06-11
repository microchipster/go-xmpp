package xmpp

import (
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
