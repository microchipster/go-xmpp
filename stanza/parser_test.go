// SPDX-FileCopyrightText: 2026 Slavi Pantaleev
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package stanza

import (
	"encoding/xml"
	"strings"
	"testing"
)

func TestNextStartTreatsClosedXMLStreamAsConnectionClosed(t *testing.T) {
	for _, input := range []string{"", "<challenge"} {
		_, err := NextStart(xml.NewDecoder(strings.NewReader(input)))
		if err == nil || err.Error() != "connection closed" {
			t.Fatalf("NextStart(%q) error = %v, want connection closed", input, err)
		}
	}
}

func TestNextXmppTokenTreatsClosedXMLStreamAsConnectionClosed(t *testing.T) {
	_, err := NextXmppToken(xml.NewDecoder(strings.NewReader("<stream:stream")))
	if err == nil || err.Error() != "connection closed" {
		t.Fatalf("NextXmppToken error = %v, want connection closed", err)
	}
}
