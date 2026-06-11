package xmpp

import (
	"fmt"

	"gosrc.io/xmpp/stanza"
)

func streamLangAttr(lang string) string {
	if lang == "" {
		return ""
	}
	return fmt.Sprintf(" xml:lang='%s'", lang)
}

func clientStreamOpen(domain, lang string) string {
	return fmt.Sprintf("<?xml version='1.0'?><stream:stream to='%s'%s xmlns='%s' xmlns:stream='%s' version='1.0'>", domain, streamLangAttr(lang), stanza.NSClient, stanza.NSStream)
}

func componentStreamOpen(domain, lang string) string {
	return fmt.Sprintf("<?xml version='1.0'?><stream:stream to='%s'%s xmlns='%s' xmlns:stream='%s'>", domain, streamLangAttr(lang), stanza.NSComponent, stanza.NSStream)
}

func websocketStreamOpen(domain, lang string) string {
	return fmt.Sprintf(`<open xmlns="urn:ietf:params:xml:ns:xmpp-framing" to="%s"%s version="1.0" />`, domain, streamLangAttr(lang))
}
