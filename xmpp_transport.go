package xmpp

import (
	"bufio"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"gosrc.io/xmpp/stanza"
)

// XMPPTransport implements the XMPP native TCP transport
// The decoder is expected to be initialized after connecting to a server.
type XMPPTransport struct {
	openStatement string
	Config        TransportConfiguration
	TLSConfig     *tls.Config
	deadline      time.Time
	decoder       *xml.Decoder
	conn          net.Conn
	readWriter    io.ReadWriter
	logFile       io.Writer
	isSecure      bool
	// Used to close TCP connection when a stream close message is received from the server
	closeChan chan stanza.StreamClosePacket
}

var componentStreamOpen = fmt.Sprintf("<?xml version='1.0'?><stream:stream to='%%s' xmlns='%s' xmlns:stream='%s'>", stanza.NSComponent, stanza.NSStream)

var clientStreamOpen = fmt.Sprintf("<?xml version='1.0'?><stream:stream to='%%s' xmlns='%s' xmlns:stream='%s' version='1.0'>", stanza.NSClient, stanza.NSStream)

func (t *XMPPTransport) Connect() (string, error) {
	var err error

	t.conn, err = net.DialTimeout("tcp", t.Config.Address, time.Duration(t.Config.ConnectTimeout)*time.Second)
	if err != nil {
		return "", NewConnError(err, true)
	}

	t.closeChan = make(chan stanza.StreamClosePacket, 1)
	t.readWriter = newStreamLogger(t.conn, t.logFile)
	t.decoder = xml.NewDecoder(bufio.NewReaderSize(t.readWriter, maxPacketSize))
	t.decoder.CharsetReader = t.Config.CharsetReader
	if !t.deadline.IsZero() {
		if err := t.conn.SetDeadline(t.deadline); err != nil {
			return "", NewConnError(err, true)
		}
	}
	return t.StartStream()
}

func (t *XMPPTransport) StartStream() (string, error) {
	streamOpen := fmt.Sprintf(t.openStatement, t.Config.Domain)
	if _, err := t.Write([]byte(streamOpen)); err != nil {
		t.Close()
		return "", NewConnError(fmt.Errorf("write stream open: %w", err), true)
	}

	sessionID, err := stanza.InitStream(t.GetDecoder())
	if err != nil {
		t.Close()
		return "", NewConnError(fmt.Errorf("read stream open: %w", err), false)
	}
	return sessionID, nil
}

func (t *XMPPTransport) DoesStartTLS() bool {
	return true
}

func (t *XMPPTransport) GetDomain() string {
	return t.Config.Domain
}

func (t *XMPPTransport) GetDecoder() *xml.Decoder {
	if t.decoder == nil {
		t.decoder = newEmptyDecoder()
	}
	return t.decoder
}

func (t *XMPPTransport) IsSecure() bool {
	return t.isSecure
}

func (t *XMPPTransport) StartTLS() error {
	if t.Config.TLSConfig == nil {
		t.TLSConfig = &tls.Config{}
	} else {
		t.TLSConfig = t.Config.TLSConfig.Clone()
	}

	if t.TLSConfig.ServerName == "" {
		t.TLSConfig.ServerName = t.Config.Domain
	}
	tlsConn := tls.Client(t.conn, t.TLSConfig)
	// We convert existing connection to TLS
	if err := tlsConn.Handshake(); err != nil {
		return err
	}

	t.isSecure = false
	t.conn = tlsConn
	t.readWriter = newStreamLogger(tlsConn, t.logFile)
	t.decoder = xml.NewDecoder(bufio.NewReaderSize(t.readWriter, maxPacketSize))
	t.decoder.CharsetReader = t.Config.CharsetReader
	if !t.deadline.IsZero() {
		if err := t.conn.SetDeadline(t.deadline); err != nil {
			return err
		}
	}

	if !t.TLSConfig.InsecureSkipVerify {
		if err := tlsConn.VerifyHostname(t.Config.Domain); err != nil {
			return err
		}
	}

	t.isSecure = true
	return nil
}

func (t *XMPPTransport) SCRAMChannelBindingData(types []string) (string, []byte, error) {
	if !t.isSecure || len(types) == 0 {
		return "", nil, nil
	}
	tlsConn, ok := t.conn.(*tls.Conn)
	if !ok {
		return "", nil, nil
	}
	state := tlsConn.ConnectionState()
	if containsString(types, "tls-exporter") && state.Version == tls.VersionTLS13 {
		data, err := state.ExportKeyingMaterial("EXPORTER-Channel-Binding", nil, 32)
		if err != nil {
			return "", nil, err
		}
		return "tls-exporter", data, nil
	}
	if containsString(types, "tls-unique") && state.Version >= tls.VersionTLS10 && state.Version <= tls.VersionTLS12 && len(state.TLSUnique) > 0 {
		return "tls-unique", state.TLSUnique, nil
	}
	if containsString(types, "tls-server-end-point") {
		data, err := tlsServerEndPointData(state)
		if err != nil {
			return "", nil, err
		}
		return "tls-server-end-point", data, nil
	}
	return "", nil, nil
}

func tlsServerEndPointData(state tls.ConnectionState) ([]byte, error) {
	if len(state.PeerCertificates) == 0 {
		return nil, errors.New("channel binding: missing peer certificate")
	}
	hashFunc := tlsServerEndPointHash(state.PeerCertificates[0])
	if !hashFunc.Available() {
		return nil, fmt.Errorf("channel binding: hash %v unavailable", hashFunc)
	}
	h := hashFunc.New()
	_, _ = h.Write(state.PeerCertificates[0].Raw)
	return h.Sum(nil), nil
}

func tlsServerEndPointHash(cert *x509.Certificate) crypto.Hash {
	switch cert.SignatureAlgorithm {
	case x509.MD2WithRSA, x509.MD5WithRSA, x509.SHA1WithRSA, x509.DSAWithSHA1, x509.ECDSAWithSHA1:
		return crypto.SHA256
	case x509.SHA256WithRSA, x509.DSAWithSHA256, x509.ECDSAWithSHA256, x509.SHA256WithRSAPSS, x509.PureEd25519:
		return crypto.SHA256
	case x509.SHA384WithRSA, x509.ECDSAWithSHA384, x509.SHA384WithRSAPSS:
		return crypto.SHA384
	case x509.SHA512WithRSA, x509.ECDSAWithSHA512, x509.SHA512WithRSAPSS:
		return crypto.SHA512
	default:
		return crypto.SHA256
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func (t *XMPPTransport) Ping() error {
	n, err := t.conn.Write([]byte("\n"))
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("could not write ping")
	}
	return nil
}

func (t *XMPPTransport) SetDeadline(deadline time.Time) error {
	t.deadline = deadline
	if t.conn != nil {
		return t.conn.SetDeadline(deadline)
	}
	return nil
}

func (t *XMPPTransport) Read(p []byte) (n int, err error) {
	if t.readWriter == nil {
		return 0, errors.New("cannot read: not connected, no readwriter")
	}
	n, err = t.readWriter.Read(p)
	if err != nil {
		return n, fmt.Errorf("tcp read: %w", err)
	}
	return n, nil
}

func (t *XMPPTransport) Write(p []byte) (n int, err error) {
	if t.readWriter == nil {
		return 0, errors.New("cannot write: not connected, no readwriter")
	}
	n, err = t.readWriter.Write(p)
	if err != nil {
		return n, fmt.Errorf("tcp write via %T: %w", t.readWriter, err)
	}
	return n, nil
}

func (t *XMPPTransport) Close() error {
	if t.readWriter != nil {
		_, _ = t.readWriter.Write([]byte(stanza.StreamClose))
	}

	// Try to wait for the stream close tag from the server. After a timeout, disconnect anyway.
	if t.closeChan != nil {
		select {
		case <-t.closeChan:
		case <-time.After(time.Duration(t.Config.ConnectTimeout) * time.Second):
		}
	}

	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}

func (t *XMPPTransport) LogTraffic(logFile io.Writer) {
	t.logFile = logFile
}

func (t *XMPPTransport) ReceivedStreamClose() {
	if t.closeChan == nil {
		return
	}
	select {
	case t.closeChan <- stanza.StreamClosePacket{}:
	default:
	}
}
