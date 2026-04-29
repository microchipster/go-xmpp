package xmpp

import "testing"

func TestParseSCRAMFieldsRejectsDuplicateAttributes(t *testing.T) {
	if _, err := parseSCRAMFields("r=nonce,s=salt,r=again"); err == nil {
		t.Fatal("expected duplicate SCRAM attribute to fail")
	}
}

func TestSCRAMDowngradeString(t *testing.T) {
	got, err := scramDowngradeString(
		[]string{"PLAIN", "SCRAM-SHA-512", "SCRAM-SHA-1"},
		[]string{"tls-exporter", "tls-server-end-point"},
	)
	if err != nil {
		t.Fatalf("scramDowngradeString returned error: %v", err)
	}
	want := "PLAIN\x1eSCRAM-SHA-1\x1eSCRAM-SHA-512\x1ftls-exporter\x1etls-server-end-point"
	if got != want {
		t.Fatalf("scramDowngradeString() = %q, want %q", got, want)
	}
}

func TestSCRAMGS2Header(t *testing.T) {
	if got := scramGS2Header(""); got != "y,," {
		t.Fatalf("expected non-PLUS SCRAM GS2 header, got %q", got)
	}
	if got := scramGS2Header("tls-exporter"); got != "p=tls-exporter,," {
		t.Fatalf("expected SCRAM-PLUS GS2 header, got %q", got)
	}
}

func TestSelectSASLMechanismPrefersPlusWithChannelBinding(t *testing.T) {
	preferred := Password("secret").mechanisms
	advertised := []string{"PLAIN", "SCRAM-SHA-512", "SCRAM-SHA-1-PLUS"}
	if got := selectSASLMechanism(preferred, advertised, "tls-exporter"); got != "SCRAM-SHA-1-PLUS" {
		t.Fatalf("selectSASLMechanism() = %q, want SCRAM-SHA-1-PLUS", got)
	}
}

func TestSelectSASLMechanismFallsBackWithoutChannelBinding(t *testing.T) {
	preferred := Password("secret").mechanisms
	advertised := []string{"PLAIN", "SCRAM-SHA-512-PLUS", "SCRAM-SHA-512"}
	if got := selectSASLMechanism(preferred, advertised, ""); got != "SCRAM-SHA-512" {
		t.Fatalf("selectSASLMechanism() = %q, want SCRAM-SHA-512", got)
	}
}

func TestSCRAMDowngradeStringRejectsInvalidNames(t *testing.T) {
	if _, err := scramDowngradeString([]string{"plain"}, nil); err == nil {
		t.Fatal("expected invalid SASL mechanism name to fail")
	}
	if _, err := scramDowngradeString([]string{"PLAIN"}, []string{"tls_exporter"}); err == nil {
		t.Fatal("expected invalid channel binding name to fail")
	}
}
