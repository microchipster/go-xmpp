package xmpp

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"gosrc.io/xmpp/stanza"
)

const (
	// Default port is not standard XMPP port to avoid interfering
	// with local running XMPP server
	testXMPPAddress  = "localhost:15222"
	testClientDomain = "localhost"
)

func TestEventManager(t *testing.T) {
	mgr := EventManager{}
	mgr.updateState(StateResuming)
	if mgr.CurrentState.getState() != StateResuming {
		t.Fatal("CurrentState not updated by updateState()")
	}

	mgr.disconnected(SMState{})

	if mgr.CurrentState.getState() != StateDisconnected {
		t.Fatalf("CurrentState not reset by disconnected()")
	}

	mgr.streamError(ErrTLSNotSupported.Error(), "")

	if mgr.CurrentState.getState() != StateStreamError {
		t.Fatalf("CurrentState not set by streamError()")
	}
}

func TestClient_Connect(t *testing.T) {
	// Setup Mock server
	mock := ServerMock{}
	mock.Start(t, testXMPPAddress, handlerClientConnectSuccess)

	// Test / Check result
	config := Config{
		TransportConfiguration: TransportConfiguration{
			Address: testXMPPAddress,
		},
		Jid:        "test@localhost",
		Credential: Password("test"),
		Insecure:   true}

	var client *Client
	var err error
	router := NewRouter()
	if client, err = NewClient(&config, router, clientDefaultErrorHandler); err != nil {
		t.Errorf("connect create XMPP client: %s", err)
	}

	if err = client.Connect(); err != nil {
		t.Errorf("XMPP connection failed: %s", err)
	}

	mock.Stop()
}

func TestClient_NoInsecure(t *testing.T) {
	// Setup Mock server
	mock := ServerMock{}
	mock.Start(t, testXMPPAddress, func(t *testing.T, sc *ServerConn) {
		handlerAbortTLS(t, sc)
		closeConn(t, sc)
	})

	// Test / Check result
	config := Config{
		TransportConfiguration: TransportConfiguration{
			Address: testXMPPAddress,
		},
		Jid:        "test@localhost",
		Credential: Password("test"),
	}

	var client *Client
	var err error
	router := NewRouter()
	if client, err = NewClient(&config, router, clientDefaultErrorHandler); err != nil {
		t.Errorf("cannot create XMPP client: %s", err)
	}

	if err = client.Connect(); err == nil {
		// When insecure is not allowed:
		t.Errorf("should fail as insecure connection is not allowed and server does not support TLS")
	}

	mock.Stop()
}

func TestClient_SessionEstablishmentTimeout(t *testing.T) {
	mock := ServerMock{}
	mock.Start(t, testXMPPAddress, handlerClientSessionTimeout)
	defer mock.Stop()

	config := Config{
		TransportConfiguration: TransportConfiguration{
			Address: testXMPPAddress,
		},
		Jid:            "test@localhost",
		Credential:     Password("test"),
		Insecure:       true,
		ConnectTimeout: 1,
	}

	var client *Client
	var err error
	router := NewRouter()
	if client, err = NewClient(&config, router, clientDefaultErrorHandler); err != nil {
		t.Fatalf("cannot create XMPP client: %s", err)
	}

	start := time.Now()
	err = client.Connect()
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("expected session establishment to time out")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "timeout") {
		t.Fatalf("expected timeout error, got %v", err)
	}
	if elapsed > 4*time.Second {
		t.Fatalf("expected session establishment timeout to fail quickly, took %s", elapsed)
	}
}

func TestClient_SRVFallback(t *testing.T) {
	originalLookupSRV := lookupSRV
	lookupSRV = func(service, proto, name string) (string, []*net.SRV, error) {
		if service != "xmpp-client" || proto != "tcp" || name != testClientDomain {
			t.Fatalf("unexpected SRV lookup: %s %s %s", service, proto, name)
		}
		return "", []*net.SRV{
			{Target: "localhost", Port: testClientSrvFallbackDead, Priority: 10, Weight: 20},
			{Target: "localhost", Port: testClientSrvFallbackSuccess, Priority: 10, Weight: 10},
		}, nil
	}
	defer func() { lookupSRV = originalLookupSRV }()

	mock := ServerMock{}
	mock.Start(t, fmt.Sprintf("localhost:%d", testClientSrvFallbackSuccess), handlerClientConnectSuccess)
	defer mock.Stop()

	config := Config{
		Jid:            "test@localhost",
		Credential:     Password("test"),
		Insecure:       true,
		ConnectTimeout: 2,
	}

	var client *Client
	var err error
	router := NewRouter()
	if client, err = NewClient(&config, router, clientDefaultErrorHandler); err != nil {
		t.Fatalf("cannot create XMPP client: %s", err)
	}

	if got, want := client.config.Address, fmt.Sprintf("localhost:%d", testClientSrvFallbackDead); got != want {
		t.Fatalf("expected first SRV target to be selected initially, got %q want %q", got, want)
	}

	if err = client.Connect(); err != nil {
		t.Fatalf("XMPP connection failed through SRV fallback: %s", err)
	}

	if client.transport == nil {
		t.Fatal("expected transport to be established")
	}
	if client.config.Address != fmt.Sprintf("localhost:%d", testClientSrvFallbackSuccess) {
		t.Fatalf("expected fallback to second SRV target, got %q", client.config.Address)
	}

	if err := client.Disconnect(); err != nil {
		t.Fatalf("disconnect after SRV fallback failed: %v", err)
	}
}

// Check that the client is properly tracking features, as session negotiation progresses.
func TestClient_FeaturesTracking(t *testing.T) {
	serverDone := make(chan struct{})
	// Setup Mock server
	mock := ServerMock{}
	mock.Start(t, testXMPPAddress, func(t *testing.T, sc *ServerConn) {
		checkClientOpenStream(t, sc)

		sendStreamFeaturesWithCaps(t, sc)
		readAuth(t, sc.decoder)
		sc.connection.Write([]byte("<success xmlns=\"urn:ietf:params:xml:ns:xmpp-sasl\"/>"))

		checkClientOpenStream(t, sc)
		sendBindFeature(t, sc)
		bind(t, sc)
		serverDone <- struct{}{}
	})

	// Test / Check result
	config := Config{
		TransportConfiguration: TransportConfiguration{
			Address: testXMPPAddress,
		},
		Jid:        "test@localhost",
		Credential: Password("test"),
		Insecure:   true,
	}

	var client *Client
	var err error
	router := NewRouter()
	if client, err = NewClient(&config, router, clientDefaultErrorHandler); err != nil {
		t.Errorf("cannot create XMPP client: %s", err)
	}

	if err = client.Connect(); err != nil {
		t.Fatalf("could not connect client to mock server: %s", err)
	}

	serverFeatures := client.ServerFeatures()
	if serverFeatures.Caps.Ver != "server-cap-v1" {
		t.Fatalf("expected pre-auth server caps to be tracked, got %#v", serverFeatures.Caps)
	}
	if len(serverFeatures.Mechanisms.Mechanism) != 1 || serverFeatures.Mechanisms.Mechanism[0] != "PLAIN" {
		t.Fatalf("expected initial server mechanisms to be tracked, got %#v", serverFeatures.Mechanisms.Mechanism)
	}

	if client.Session == nil {
		t.Fatal("expected session to be established")
	}
	if client.Session.Features.Bind.XMLName.Space != stanza.NSBind {
		t.Fatalf("expected post-auth features to be tracked, got %#v", client.Session.Features.Bind)
	}
	if client.Session.ServerFeatures.Caps.Ver != "server-cap-v1" {
		t.Fatalf("expected session server features to remain available, got %#v", client.Session.ServerFeatures.Caps)
	}

	select {
	case <-serverDone:
	case <-time.After(defaultChannelTimeout):
		t.Fatal("mock server did not finish feature tracking handshake")
	}

	mock.Stop()
}

func TestClient_ServerCapabilitiesAreCached(t *testing.T) {
	serverDone := make(chan struct{})
	// Setup Mock server
	mock := ServerMock{}
	mock.Start(t, testXMPPAddress, func(t *testing.T, sc *ServerConn) {
		checkClientOpenStream(t, sc)

		sendStreamFeaturesWithCaps(t, sc)
		readAuth(t, sc.decoder)
		sc.connection.Write([]byte("<success xmlns=\"urn:ietf:params:xml:ns:xmpp-sasl\"/>"))

		checkClientOpenStream(t, sc)
		sendBindFeature(t, sc)
		bind(t, sc)
		discardPresence(t, sc)
		respondToIQ(t, sc)
		serverDone <- struct{}{}
	})

	config := Config{
		TransportConfiguration: TransportConfiguration{
			Address: testXMPPAddress,
		},
		Jid:        "test@localhost",
		Credential: Password("test"),
		Insecure:   true,
	}

	var client *Client
	var err error
	router := NewRouter()
	if client, err = NewClient(&config, router, clientDefaultErrorHandler); err != nil {
		t.Fatalf("cannot create XMPP client: %s", err)
	}

	if err = client.Connect(); err != nil {
		t.Fatalf("XMPP connection failed: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, err := client.ServerCapabilities(ctx)
	if err != nil {
		t.Fatalf("failed to fetch server capabilities: %v", err)
	}
	if !info.HasFeature("vcard-temp") {
		t.Fatalf("expected server disco info to include vcard-temp, got %#v", info.Features)
	}
	if !info.HasFeature("http://jabber.org/protocol/address") {
		t.Fatalf("expected server disco info to include address feature, got %#v", info.Features)
	}

	select {
	case <-serverDone:
	case <-time.After(defaultChannelTimeout):
		t.Fatal("mock server did not answer the initial capabilities request")
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel2()

	info2, err := client.ServerCapabilities(ctx2)
	if err != nil {
		t.Fatalf("expected cached capabilities lookup to succeed, got %v", err)
	}
	if !info2.HasFeature("vcard-temp") || !info2.HasFeature("http://jabber.org/protocol/address") {
		t.Fatalf("cached capabilities lost features: %#v", info2.Features)
	}

	mock.Stop()
}

func TestClient_RFC3921Session(t *testing.T) {
	// Setup Mock server
	mock := ServerMock{}
	mock.Start(t, testXMPPAddress, handlerClientConnectWithSession)

	// Test / Check result
	config := Config{
		TransportConfiguration: TransportConfiguration{
			Address: testXMPPAddress,
		},
		Jid:        "test@localhost",
		Credential: Password("test"),
		Insecure:   true,
	}

	var client *Client
	var err error
	router := NewRouter()
	if client, err = NewClient(&config, router, clientDefaultErrorHandler); err != nil {
		t.Errorf("connect create XMPP client: %s", err)
	}

	if err = client.Connect(); err != nil {
		t.Errorf("XMPP connection failed: %s", err)
	}

	mock.Stop()
}

// Testing sending an IQ to the mock server and reading its response.
func TestClient_SendIQ(t *testing.T) {
	done := make(chan struct{})
	// Handler for Mock server
	h := func(t *testing.T, sc *ServerConn) {
		handlerClientConnectSuccess(t, sc)
		discardPresence(t, sc)
		respondToIQ(t, sc)
		done <- struct{}{}
	}
	client, mock := mockClientConnection(t, h, testClientIqPort)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	iqReq, err := stanza.NewIQ(stanza.Attrs{Type: stanza.IQTypeGet, From: "test1@localhost/mremond-mbp", To: defaultServerName, Id: defaultStreamID, Lang: "en"})
	if err != nil {
		t.Fatalf("failed to create the IQ request: %v", err)
	}

	disco := iqReq.DiscoInfo()
	iqReq.Payload = disco

	// Handle a possible error
	errChan := make(chan error)
	errorHandler := func(err error) {
		errChan <- err
	}
	client.ErrorHandler = errorHandler
	res, err := client.SendIQ(ctx, iqReq)
	if err != nil {
		t.Error(err)
	}

	select {
	case <-res: // If the server responds with an IQ, we pass the test
	case err := <-errChan: // If the server sends an error, or there is a connection error
		cancel()
		t.Fatal(err.Error())
	case <-time.After(defaultChannelTimeout): // If we timeout
		cancel()
		t.Fatal("Failed to receive response, to sent IQ, from mock server")
	}
	select {
	case <-done:
		mock.Stop()
	case <-time.After(defaultChannelTimeout):
		cancel()
		t.Fatal("The mock server failed to finish its job !")
	}
	cancel()
}

func TestClient_SendIQFail(t *testing.T) {
	done := make(chan struct{})
	// Handler for Mock server
	h := func(t *testing.T, sc *ServerConn) {
		handlerClientConnectSuccess(t, sc)
		discardPresence(t, sc)
		respondToIQ(t, sc)
		done <- struct{}{}
	}
	client, mock := mockClientConnection(t, h, testClientIqFailPort)

	//==================
	// Create an IQ to send
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	iqReq, err := stanza.NewIQ(stanza.Attrs{Type: stanza.IQTypeGet, From: "test1@localhost/mremond-mbp", To: defaultServerName, Id: defaultStreamID, Lang: "en"})
	if err != nil {
		t.Fatalf("failed to create IQ request: %v", err)
	}
	disco := iqReq.DiscoInfo()
	iqReq.Payload = disco
	// Removing the id to make the stanza invalid. The IQ constructor makes a random one if none is specified
	// so we need to overwrite it.
	iqReq.Id = ""

	// Handle a possible error
	errChan := make(chan error)
	errorHandler := func(err error) {
		errChan <- err
	}
	client.ErrorHandler = errorHandler
	res, _ := client.SendIQ(ctx, iqReq)

	// Test
	select {
	case <-res: // If the server responds with an IQ
		t.Errorf("Server should not respond with an IQ since the request is expected to be invalid !")
	case <-errChan: // If the server sends an error, the test passes
	case <-time.After(defaultChannelTimeout): // If we timeout
		t.Errorf("Failed to receive response, to sent IQ, from mock server")
	}
	select {
	case <-done:
		mock.Stop()
	case <-time.After(defaultChannelTimeout):
		cancel()
		t.Errorf("The mock server failed to finish its job !")
	}
	cancel()
}

func TestClient_SendRaw(t *testing.T) {
	done := make(chan struct{})
	// Handler for Mock server
	h := func(t *testing.T, sc *ServerConn) {
		handlerClientConnectSuccess(t, sc)
		discardPresence(t, sc)
		respondToIQ(t, sc)
		closeConn(t, sc)
		done <- struct{}{}
	}
	type testCase struct {
		req       string
		shouldErr bool
		port      int
	}
	testRequests := make(map[string]testCase)
	// Sending a correct IQ of type get. Not supposed to err
	testRequests["Correct IQ"] = testCase{
		req:       `<iq type="get" id="91bd0bba-012f-4d92-bb17-5fc41e6fe545" from="test1@localhost/mremond-mbp" to="testServer" lang="en"><query xmlns="http://jabber.org/protocol/disco#info"></query></iq>`,
		shouldErr: false,
		port:      testClientRawPort + 100,
	}
	// Sending an IQ with a missing ID. Should err
	testRequests["IQ with missing ID"] = testCase{
		req:       `<iq type="get" from="test1@localhost/mremond-mbp" to="testServer" lang="en"><query xmlns="http://jabber.org/protocol/disco#info"></query></iq>`,
		shouldErr: true,
		port:      testClientRawPort,
	}

	// A handler for the client.
	// In the failing test, the server returns a stream error, which triggers this handler, client side.
	errChan := make(chan error)
	errHandler := func(err error) {
		errChan <- err
	}

	// Tests for all the IQs
	for name, tcase := range testRequests {
		t.Run(name, func(st *testing.T) {
			//Connecting to a mock server, initialized with given port and handler function
			c, m := mockClientConnection(t, h, tcase.port)
			c.ErrorHandler = errHandler
			// Sending raw xml from test case
			err := c.SendRaw(tcase.req)
			if err != nil {
				t.Errorf("Error sending Raw string")
			}
			// Just wait a little so the message has time to arrive
			select {
			// We don't use the default "long" timeout here because waiting it out means passing the test.
			case <-time.After(100 * time.Millisecond):
				c.Disconnect()
			case err = <-errChan:
				if err == nil && tcase.shouldErr {
					t.Errorf("Failed to get closing stream err")
				} else if err != nil && !tcase.shouldErr {
					t.Errorf("This test is not supposed to err !")
				}
			}
			select {
			case <-done:
				m.Stop()
			case <-time.After(defaultChannelTimeout):
				t.Errorf("The mock server failed to finish its job !")
			}
		})
	}
}

func TestClient_Disconnect(t *testing.T) {
	c, m := mockClientConnection(t, func(t *testing.T, sc *ServerConn) {
		handlerClientConnectSuccess(t, sc)
		closeConn(t, sc)
	}, testClientBasePort)
	err := c.transport.Ping()
	if err != nil {
		t.Errorf("Could not ping but not disconnected yet")
	}
	c.Disconnect()
	if c.CurrentState.getState() != StateDisconnected {
		t.Errorf("Did not update state to disconnected")
	}
	err = c.transport.Ping()
	if err == nil {
		t.Errorf("Did not disconnect properly")
	}
	m.Stop()
}

func TestClient_DisconnectStreamManager(t *testing.T) {
	// Init mock server
	// Setup Mock server
	mock := ServerMock{}
	mock.Start(t, testXMPPAddress, func(t *testing.T, sc *ServerConn) {
		handlerAbortTLS(t, sc)
		closeConn(t, sc)
	})

	// Test / Check result
	config := Config{
		TransportConfiguration: TransportConfiguration{
			Address: testXMPPAddress,
		},
		Jid:        "test@localhost",
		Credential: Password("test"),
	}

	var client *Client
	var err error
	router := NewRouter()
	if client, err = NewClient(&config, router, clientDefaultErrorHandler); err != nil {
		t.Errorf("cannot create XMPP client: %s", err)
	}

	sman := NewStreamManager(client, nil)
	errChan := make(chan error)
	runSMan := func(errChan chan error) {
		errChan <- sman.Run()
	}

	go runSMan(errChan)
	select {
	case <-errChan:
	case <-time.After(defaultChannelTimeout):
		// When insecure is not allowed:
		t.Errorf("should fail as insecure connection is not allowed and server does not support TLS")
	}
	mock.Stop()
}

func Test_ClientPostConnectHook(t *testing.T) {
	done := make(chan struct{})
	// Handler for Mock server
	h := func(t *testing.T, sc *ServerConn) {
		handlerClientConnectSuccess(t, sc)
		done <- struct{}{}
	}

	hookChan := make(chan struct{})
	mock := &ServerMock{}
	testServerAddress := fmt.Sprintf("%s:%d", testClientDomain, testClientPostConnectHook)

	mock.Start(t, testServerAddress, h)
	config := Config{
		TransportConfiguration: TransportConfiguration{
			Address: testServerAddress,
		},
		Jid:        "test@localhost",
		Credential: Password("test"),
		Insecure:   true}

	var client *Client
	var err error
	router := NewRouter()
	if client, err = NewClient(&config, router, clientDefaultErrorHandler); err != nil {
		t.Errorf("connect create XMPP client: %s", err)
	}

	// The post connection client hook should just write to a channel that we will read later.
	client.PostConnectHook = func() error {
		go func() {
			hookChan <- struct{}{}
		}()
		return nil
	}
	// Handle a possible error
	errChan := make(chan error)
	errorHandler := func(err error) {
		errChan <- err
	}
	client.ErrorHandler = errorHandler
	if err = client.Connect(); err != nil {
		t.Errorf("XMPP connection failed: %s", err)
	}

	// Check if the post connection client hook was correctly called
	select {
	case err := <-errChan: // If the server sends an error, or there is a connection error
		t.Fatal(err.Error())
	case <-time.After(defaultChannelTimeout): // If we timeout
		t.Fatal("Failed to call post connection client hook")
	case <-hookChan:
		// Test succeeded, channel was written to.
	}

	select {
	case <-done:
		mock.Stop()
	case <-time.After(defaultChannelTimeout):
		t.Fatal("The mock server failed to finish its job !")
	}
}

func Test_ClientPostReconnectHook(t *testing.T) {
	hookChan := make(chan struct{})
	// Setup Mock server
	mock := ServerMock{}
	mock.Start(t, testXMPPAddress, func(t *testing.T, sc *ServerConn) {
		checkClientOpenStream(t, sc)

		sendStreamFeatures(t, sc) // Send initial features
		readAuth(t, sc.decoder)
		sc.connection.Write([]byte("<success xmlns=\"urn:ietf:params:xml:ns:xmpp-sasl\"/>"))

		checkClientOpenStream(t, sc)       // Reset stream
		sendFeaturesStreamManagment(t, sc) // Send post auth features
		bind(t, sc)
		enableStreamManagement(t, sc, false, true)
	})

	// Test / Check result
	config := Config{
		TransportConfiguration: TransportConfiguration{
			Address: testXMPPAddress,
		},
		Jid:                    "test@localhost",
		Credential:             Password("test"),
		Insecure:               true,
		StreamManagementEnable: true,
		StreamManagementResume: true} // Enable stream management

	var client *Client
	router := NewRouter()
	client, err := NewClient(&config, router, clientDefaultErrorHandler)
	if err != nil {
		t.Errorf("connect create XMPP client: %s", err)
	}

	client.PostResumeHook = func() error {
		go func() {
			hookChan <- struct{}{}
		}()
		return nil
	}

	err = client.Connect()
	if err != nil {
		t.Fatalf("could not connect client to mock server: %s", err)
	}

	transp, ok := client.transport.(*XMPPTransport)
	if !ok {
		t.Fatalf("problem with client transport ")
	}

	transp.conn.Close()
	mock.Stop()

	// Check if the client can have its connection resumed using its state but also its configuration
	if !IsStreamResumable(client) {
		t.Fatalf("should support resumption")
	}

	// Reboot server. We need to make a new one because (at least for now) the mock server can only have one handler
	// and they should be different between a first connection and a stream resume since exchanged messages
	// are different (See XEP-0198)
	mock2 := ServerMock{}
	mock2.Start(t, testXMPPAddress, func(t *testing.T, sc *ServerConn) {
		//	Reconnect
		checkClientOpenStream(t, sc)

		sendStreamFeatures(t, sc) // Send initial features
		readAuth(t, sc.decoder)
		sc.connection.Write([]byte("<success xmlns=\"urn:ietf:params:xml:ns:xmpp-sasl\"/>"))

		checkClientOpenStream(t, sc)       // Reset stream
		sendFeaturesStreamManagment(t, sc) // Send post auth features
		resumeStream(t, sc)
	})

	// Reconnect
	err = client.Resume()
	if err != nil {
		t.Fatalf("could not connect client to mock server: %s", err)
	}

	select {
	case <-time.After(defaultChannelTimeout): // If we timeout
		t.Fatal("Failed to call post connection client hook")
	case <-hookChan:
		// Test succeeded, channel was written to.
	}

	mock2.Stop()
}

//=============================================================================
// Basic XMPP Server Mock Handlers.

// Test connection with a basic straightforward workflow
func handlerClientConnectSuccess(t *testing.T, sc *ServerConn) {
	checkClientOpenStream(t, sc)
	sendStreamFeatures(t, sc) // Send initial features
	readAuth(t, sc.decoder)
	sc.connection.Write([]byte("<success xmlns=\"urn:ietf:params:xml:ns:xmpp-sasl\"/>"))

	checkClientOpenStream(t, sc) // Reset stream
	sendBindFeature(t, sc)       // Send post auth features
	bind(t, sc)
}

// closeConn closes the connection on request from the client
func closeConn(t *testing.T, sc *ServerConn) {
	for {
		cls, err := stanza.NextPacket(sc.decoder)
		if err != nil {
			if err == io.EOF || errors.Is(err, net.ErrClosed) || strings.Contains(err.Error(), "closed network connection") || strings.Contains(err.Error(), "connection reset by peer") {
				return
			}
			t.Errorf("cannot read from socket: %s", err)
			return
		}
		switch cls.(type) {
		case stanza.StreamClosePacket:
			sc.connection.Write([]byte(stanza.StreamClose))
			return
		}
	}

}

// We expect client will abort on TLS
func handlerAbortTLS(t *testing.T, sc *ServerConn) {
	checkClientOpenStream(t, sc)
	sendStreamFeatures(t, sc) // Send initial features
}

func handlerClientSessionTimeout(t *testing.T, sc *ServerConn) {
	checkClientOpenStream(t, sc)
	sendStreamFeatures(t, sc)
	time.Sleep(3 * time.Second)
}

// Test connection with mandatory session (RFC-3921)
func handlerClientConnectWithSession(t *testing.T, sc *ServerConn) {
	checkClientOpenStream(t, sc)

	sendStreamFeatures(t, sc) // Send initial features
	readAuth(t, sc.decoder)
	sc.connection.Write([]byte("<success xmlns=\"urn:ietf:params:xml:ns:xmpp-sasl\"/>"))

	checkClientOpenStream(t, sc) // Reset stream
	sendRFC3921Feature(t, sc)    // Send post auth features
	bind(t, sc)
	session(t, sc)
}

func checkClientOpenStream(t *testing.T, sc *ServerConn) {
	err := sc.connection.SetDeadline(time.Now().Add(defaultTimeout))
	if err != nil {
		t.Fatalf("failed to set deadline: %v", err)
	}
	defer sc.connection.SetDeadline(time.Time{})

	for { // TODO clean up. That for loop is not elegant and I prefer bounded recursion.
		var token xml.Token
		token, err := sc.decoder.Token()
		if err != nil {
			t.Fatalf("cannot read next token: %s", err)
		}

		switch elem := token.(type) {
		// Wait for first startElement
		case xml.StartElement:
			if elem.Name.Space != stanza.NSStream || elem.Name.Local != "stream" {
				err = errors.New("xmpp: expected <stream> but got <" + elem.Name.Local + "> in " + elem.Name.Space)
				return
			}
			if _, err := fmt.Fprintf(sc.connection, serverStreamOpen, "localhost", "streamid1", stanza.NSClient, stanza.NSStream); err != nil {
				t.Errorf("cannot write server stream open: %s", err)
			}
			return
		}

	}
}

func mockClientConnection(t *testing.T, serverHandler func(*testing.T, *ServerConn), port int) (*Client, *ServerMock) {
	mock := &ServerMock{}
	testServerAddress := fmt.Sprintf("%s:%d", testClientDomain, port)

	mock.Start(t, testServerAddress, serverHandler)
	config := Config{
		TransportConfiguration: TransportConfiguration{
			Address: testServerAddress,
		},
		Jid:        "test@localhost",
		Credential: Password("test"),
		Insecure:   true}

	var client *Client
	var err error
	router := NewRouter()
	if client, err = NewClient(&config, router, clientDefaultErrorHandler); err != nil {
		t.Errorf("connect create XMPP client: %s", err)
	}

	if err = client.Connect(); err != nil {
		t.Errorf("XMPP connection failed: %s", err)
	}

	return client, mock
}

// This really should not be used as is.
// It's just meant to be a placeholder when error handling is not needed at this level
func clientDefaultErrorHandler(err error) {
}
