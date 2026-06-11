package xmpp

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"testing"
	"time"

	"encoding/xml"
	"gosrc.io/xmpp/stanza"
)

const (
	testCertCheckerStartTLSPort  = 15301
	testCertCheckerDirectTLSPort = 15302
)

func TestServerCheck_CheckStartTLS(t *testing.T) {
	tlsCert, roots := testCertificateAuthority(t)
	originalSystemCertPool := systemCertPool
	systemCertPool = func() (*x509.CertPool, error) { return roots, nil }
	t.Cleanup(func() { systemCertPool = originalSystemCertPool })

	addr := fmt.Sprintf("127.0.0.1:%d", testCertCheckerStartTLSPort)
	ready := make(chan struct{})
	go func() {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			t.Errorf("failed to listen: %v", err)
			close(ready)
			return
		}
		defer ln.Close()
		close(ready)

		conn, err := ln.Accept()
		if err != nil {
			t.Errorf("failed to accept connection: %v", err)
			return
		}
		defer conn.Close()

		decoder := xml.NewDecoder(conn)
		if _, err := stanza.InitStream(decoder); err != nil {
			t.Errorf("failed to read client stream: %v", err)
			return
		}

		if _, err := fmt.Fprintf(conn, serverStreamOpen, "localhost", "streamid1", stanza.NSClient, stanza.NSStream); err != nil {
			t.Errorf("failed to send server stream open: %v", err)
			return
		}
		if _, err := fmt.Fprintln(conn, `<stream:features><starttls xmlns="urn:ietf:params:xml:ns:xmpp-tls"/></stream:features>`); err != nil {
			t.Errorf("failed to send stream features: %v", err)
			return
		}

		se, err := stanza.NextStart(decoder)
		if err != nil {
			t.Errorf("failed to read starttls: %v", err)
			return
		}
		var startTLS struct {
			XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-tls starttls"`
		}
		if err := decoder.DecodeElement(&startTLS, &se); err != nil {
			t.Errorf("failed to decode starttls: %v", err)
			return
		}

		if _, err := fmt.Fprintln(conn, `<proceed xmlns="urn:ietf:params:xml:ns:xmpp-tls"/>`); err != nil {
			t.Errorf("failed to send proceed: %v", err)
			return
		}

		tlsConn := tls.Server(conn, &tls.Config{Certificates: []tls.Certificate{tlsCert}})
		defer tlsConn.Close()
		if err := tlsConn.Handshake(); err != nil {
			t.Errorf("failed TLS handshake: %v", err)
		}
	}()
	<-ready

	checker, err := NewChecker(addr, "localhost")
	if err != nil {
		t.Fatalf("failed to create checker: %v", err)
	}

	if err := checker.Check(); err != nil {
		t.Fatalf("starttls check failed: %v", err)
	}
}

func TestServerCheck_CheckDirectTLS(t *testing.T) {
	tlsCert, roots := testCertificateAuthority(t)
	originalSystemCertPool := systemCertPool
	systemCertPool = func() (*x509.CertPool, error) { return roots, nil }
	t.Cleanup(func() { systemCertPool = originalSystemCertPool })

	addr := fmt.Sprintf("127.0.0.1:%d", testCertCheckerDirectTLSPort)
	ready := make(chan struct{})
	go func() {
		ln, err := tls.Listen("tcp", addr, &tls.Config{Certificates: []tls.Certificate{tlsCert}})
		if err != nil {
			t.Errorf("failed to listen: %v", err)
			close(ready)
			return
		}
		defer ln.Close()
		close(ready)

		conn, err := ln.Accept()
		if err != nil {
			t.Errorf("failed to accept connection: %v", err)
			return
		}
		defer conn.Close()

		tlsConn := conn.(*tls.Conn)
		if err := tlsConn.Handshake(); err != nil {
			t.Errorf("failed TLS handshake: %v", err)
		}
	}()
	<-ready

	checker, err := NewChecker(addr, "localhost")
	if err != nil {
		t.Fatalf("failed to create checker: %v", err)
	}
	checker.directTLS = true

	if err := checker.Check(); err != nil {
		t.Fatalf("direct tls check failed: %v", err)
	}
}

func testCertificateAuthority(t *testing.T) (tls.Certificate, *x509.CertPool) {
	t.Helper()

	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate CA key: %v", err)
	}
	caTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "xmpp-test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(30 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("failed to create CA certificate: %v", err)
	}
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		t.Fatalf("failed to parse CA certificate: %v", err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(caCert)

	leafKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate leaf key: %v", err)
	}
	leafTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(2),
		Subject:               pkix.Name{CommonName: "localhost"},
		DNSNames:              []string{"localhost"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(7 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	leafDER, err := x509.CreateCertificate(rand.Reader, leafTmpl, caCert, &leafKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("failed to create leaf certificate: %v", err)
	}

	leafPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: leafDER})
	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})
	keyDER := x509.MarshalPKCS1PrivateKey(leafKey)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyDER})

	tlsCert, err := tls.X509KeyPair(append(leafPEM, caPEM...), keyPEM)
	if err != nil {
		t.Fatalf("failed to create tls key pair: %v", err)
	}
	return tlsCert, roots
}
