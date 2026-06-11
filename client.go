package xmpp

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sort"
	"sync"
	"time"

	"gosrc.io/xmpp/stanza"
)

//=============================================================================
// EventManager

// SyncConnState represents the current connection state.
type SyncConnState struct {
	sync.RWMutex
	// Current state of the client. Please use the dedicated getter and setter for this field as they are thread safe.
	state ConnState
}
type ConnState = uint8

// getState is a thread-safe getter for the current state
func (scs *SyncConnState) getState() ConnState {
	var res ConnState
	scs.RLock()
	res = scs.state
	scs.RUnlock()
	return res
}

// GetState is a thread-safe getter for the current state.
func (scs *SyncConnState) GetState() ConnState {
	return scs.getState()
}

// setState is a thread-safe setter for the current
func (scs *SyncConnState) setState(cs ConnState) {
	scs.Lock()
	scs.state = cs
	scs.Unlock()
}

// This is a the list of events happening on the connection that the
// client can be notified about.
const (
	StateDisconnected ConnState = iota
	StateResuming
	StateSessionEstablished
	StateStreamError
	StatePermanentError
	DefaultInitialPresence = "<presence/>"
)

// Event is a structure use to convey event changes related to client state. This
// is for example used to notify the client when the client get disconnected.
type Event struct {
	State       SyncConnState
	Description string
	StreamError string
	SMState     SMState
}

// SMState holds Stream Management information regarding the session that can be
// used to resume session after disconnect
type SMState struct {
	// Stream Management ID
	Id string
	// Inbound stanza count
	Inbound uint

	// IP affinity
	preferredReconAddr string

	// Error
	StreamErrorGroup stanza.StanzaErrorGroup

	// Track sent stanzas
	*stanza.UnAckQueue

	// TODO Store max and timestamp, to check if we should retry resumption or not
}

// EventHandler is use to pass events about state of the connection to
// client implementation.
type EventHandler func(Event) error

type EventManager struct {
	// Store current state. Please use "getState" and "setState" to access and/or modify this.
	CurrentState SyncConnState

	// Callback used to propagate connection state changes
	Handler EventHandler
}

// updateState changes the CurrentState in the event manager. The state read is threadsafe but there is no guarantee
// regarding the triggered callback function.
func (em *EventManager) updateState(state ConnState) {
	em.CurrentState.setState(state)
	if em.Handler != nil {
		em.Handler(Event{State: em.CurrentState})
	}
}

// disconnected changes the CurrentState in the event manager to "disconnected". The state read is threadsafe but there is no guarantee
// regarding the triggered callback function.
func (em *EventManager) disconnected(state SMState) {
	em.CurrentState.setState(StateDisconnected)
	if em.Handler != nil {
		em.Handler(Event{State: em.CurrentState, SMState: state})
	}
}

// streamError changes the CurrentState in the event manager to "streamError". The state read is threadsafe but there is no guarantee
// regarding the triggered callback function.
func (em *EventManager) streamError(error, desc string) {
	em.CurrentState.setState(StateStreamError)
	if em.Handler != nil {
		em.Handler(Event{State: em.CurrentState, StreamError: error, Description: desc})
	}
}

// Client
// ============================================================================

var ErrCanOnlySendGetOrSetIq = errors.New("SendIQ can only send get and set IQ stanzas")
var lookupSRV = net.LookupSRV

// Client is the main structure used to connect as a client on an XMPP
// server.
type Client struct {
	// Store user defined options and states
	config *Config
	// Session gather data that can be accessed by users of this library
	Session   *Session
	transport Transport
	// Router is used to dispatch packets
	router *Router
	// Connection addresses are ordered by preference and used for SRV fallback.
	srvAddresses []string
	// Track discovered server capabilities for the current connection.
	serverCapabilitiesMu sync.RWMutex
	serverCapabilities   map[string]*stanza.DiscoInfo
	// Track and broadcast connection state
	EventManager
	// Handle errors from client execution
	ErrorHandler func(error)

	// Post connection hook. This will be executed on first connection
	PostConnectHook func() error

	// Post resume hook. This will be executed after the client resumes a lost connection using StreamManagement (XEP-0198)
	PostResumeHook func() error
}

/*
Setting up the client / Checking the parameters
*/

// NewClient generates a new XMPP client, based on Config passed as parameters.
// If host is not specified, the DNS SRV should be used to find the host from the domain part of the Jid.
// Default the port to 5222.
func NewClient(config *Config, r *Router, errorHandler func(error)) (c *Client, err error) {
	var candidateAddresses []string

	if config.KeepaliveInterval == 0 {
		config.KeepaliveInterval = time.Second * 30
	}
	// Parse Jid
	if config.parsedJid, err = stanza.NewJid(config.Jid); err != nil {
		err = errors.New("missing jid")
		return nil, NewConnError(err, true)
	}

	if config.Credential.secret == "" {
		err = errors.New("missing credential")
		return nil, NewConnError(err, true)
	}

	// Fallback to jid domain
	if config.Address == "" {
		config.Address = config.parsedJid.Domain

		// Fetch SRV DNS-Entries
		if srvAddresses := resolveSRVAddresses(config.parsedJid.Domain); len(srvAddresses) > 0 {
			config.Address = srvAddresses[0]
			candidateAddresses = srvAddresses
		}
	} else {
		candidateAddresses = []string{config.Address}
	}
	if len(candidateAddresses) == 0 {
		candidateAddresses = []string{config.Address}
	}
	if config.Domain == "" {
		// Fallback to jid domain
		config.Domain = config.parsedJid.Domain
	}
	if config.Lang != "" {
		config.TransportConfiguration.Lang = config.Lang
	}

	c = new(Client)
	c.config = config
	c.router = r
	c.ErrorHandler = errorHandler

	if c.config.ConnectTimeout == 0 {
		c.config.ConnectTimeout = 15 // 15 second as default
	}

	if config.TransportConfiguration.Domain == "" {
		config.TransportConfiguration.Domain = config.parsedJid.Domain
	}
	c.config.TransportConfiguration.ConnectTimeout = c.config.ConnectTimeout
	c.srvAddresses = candidateAddresses
	c.transport = newClientTransportForAddress(c.config.TransportConfiguration, candidateAddresses[0], c.config.StreamLogger)
	c.serverCapabilities = make(map[string]*stanza.DiscoInfo)

	return
}

func resolveSRVAddresses(domain string) []string {
	_, srvEntries, err := lookupSRV("xmpp-client", "tcp", domain)
	if err != nil || len(srvEntries) == 0 {
		return nil
	}

	sort.SliceStable(srvEntries, func(i, j int) bool {
		if srvEntries[i].Priority != srvEntries[j].Priority {
			return srvEntries[i].Priority < srvEntries[j].Priority
		}
		if srvEntries[i].Weight != srvEntries[j].Weight {
			return srvEntries[i].Weight > srvEntries[j].Weight
		}
		return srvEntries[i].Target < srvEntries[j].Target
	})

	addresses := make([]string, 0, len(srvEntries))
	seen := make(map[string]struct{}, len(srvEntries))
	for _, srv := range srvEntries {
		address := ensurePort(srv.Target, int(srv.Port))
		if _, ok := seen[address]; ok {
			continue
		}
		seen[address] = struct{}{}
		addresses = append(addresses, address)
	}

	return addresses
}

func newClientTransportForAddress(config TransportConfiguration, address string, logFile *os.File) Transport {
	config.Address = address
	transport := NewClientTransport(config)
	if logFile != nil {
		transport.LogTraffic(logFile)
	}
	return transport
}

func (c *Client) connectionAddresses() []string {
	addresses := c.srvAddresses
	if len(addresses) == 0 {
		if c.config != nil && c.config.Address != "" {
			addresses = []string{c.config.Address}
		}
	}
	preferred := ""
	if c.Session != nil {
		preferred = c.Session.SMState.preferredReconAddr
	}
	return prioritizeAddress(preferred, addresses)
}

func prioritizeAddress(preferred string, addresses []string) []string {
	if preferred == "" {
		return addresses
	}

	out := make([]string, 0, len(addresses)+1)
	out = append(out, preferred)
	for _, address := range addresses {
		if address == preferred {
			continue
		}
		out = append(out, address)
	}
	return out
}

func (c *Client) startRuntime() {
	// Start the keepalive go routine
	keepaliveQuit := make(chan struct{})
	go keepalive(c.transport, c.config.KeepaliveInterval, keepaliveQuit)
	// Start the receiver go routine
	go c.recv(keepaliveQuit)
}

func (c *Client) postSessionSetup() error {
	if c.Session != nil && c.Session.Resumed {
		if c.PostResumeHook != nil {
			return c.PostResumeHook()
		}
		return nil
	}
	// TODO: Do we always want to send initial presence automatically ?
	// Do we need an option to avoid that or do we rely on client to send the presence itself ?
	initialPresence := c.config.InitialPresence
	if initialPresence == "" {
		initialPresence = DefaultInitialPresence
	}
	if err := c.sendWithWriter(c.transport, []byte(initialPresence)); err != nil {
		return err
	}
	if c.PostConnectHook != nil {
		return c.PostConnectHook()
	}
	return nil
}

// Connect establishes a first time connection to a XMPP server.
// It calls the PostConnectHook
func (c *Client) Connect() error {
	err := c.connect()
	if err != nil {
		return err
	}
	err = c.postSessionSetup()
	if err != nil {
		return err
	}
	c.startRuntime()
	return nil
}

// connect establishes an actual TCP connection, based on previously defined parameters, as well as a XMPP session
func (c *Client) connect() error {
	addresses := c.connectionAddresses()
	if len(addresses) == 0 {
		return errors.New("client has no connection addresses")
	}

	originalSession := c.Session
	var previousTransport Transport
	var lastErr error
	for idx, address := range addresses {
		if idx > 0 && previousTransport != nil {
			oldTransport := previousTransport
			go oldTransport.Close()
		}

		c.config.TransportConfiguration.Address = address
		c.config.Address = address
		c.transport = newClientTransportForAddress(c.config.TransportConfiguration, address, c.config.StreamLogger)
		if c.config.ConnectTimeout > 0 {
			deadline := time.Now().Add(time.Duration(c.config.ConnectTimeout) * time.Second)
			if err := c.transport.SetDeadline(deadline); err != nil {
				lastErr = fmt.Errorf("set deadline on %s: %w", address, err)
				continue
			}
		}
		previousTransport = c.transport
		if originalSession != nil {
			sessionCopy := *originalSession
			sessionCopy.err = nil
			sessionCopy.transport = c.transport
			c.Session = &sessionCopy
		} else {
			c.Session = nil
		}

		streamId, err := c.transport.Connect()
		if err != nil {
			lastErr = fmt.Errorf("connect to %s: %w", address, err)
			continue
		}

		var state SMState
		if c.Session, err = NewSession(c, state); err != nil {
			lastErr = fmt.Errorf("session setup on %s: %w", address, err)
			failedTransport := c.transport
			go failedTransport.Close()
			c.Session = nil
			continue
		}
		if c.config.ConnectTimeout > 0 {
			_ = c.transport.SetDeadline(time.Time{})
		}
		c.Session.StreamId = streamId
		c.updateState(StateSessionEstablished)

		return nil
	}

	c.Session = originalSession
	return lastErr
}

// Resume attempts resuming  a Stream Managed session, based on the provided stream management
// state. See XEP-0198
func (c *Client) Resume() error {
	c.EventManager.updateState(StateResuming)
	err := c.connect()
	if err != nil {
		return err
	}
	err = c.postSessionSetup()
	if err != nil {
		return err
	}
	c.startRuntime()
	return nil
}

// Disconnect disconnects the client from the server, sending a stream close nonza and closing the TCP connection.
func (c *Client) Disconnect() error {
	if c.transport != nil {
		err := c.transport.Close()
		if c.Session != nil {
			c.disconnected(c.Session.SMState)
		}
		return err
	}
	// No transport so no connection.
	return nil
}

func (c *Client) SetHandler(handler EventHandler) {
	c.Handler = handler
}

// ServerFeatures returns the latest stream features advertised by the server before SASL/authentication.
func (c *Client) ServerFeatures() stanza.StreamFeatures {
	if c.Session == nil {
		return stanza.StreamFeatures{}
	}
	return c.Session.ServerFeatures
}

// ServerCapabilities fetches and caches the server disco#info response.
// If the server advertises entity capabilities, those are used as the cache key.
func (c *Client) ServerCapabilities(ctx context.Context) (*stanza.DiscoInfo, error) {
	if c.Session == nil {
		return nil, errors.New("client session is not established")
	}

	cacheKey := c.serverCapabilitiesCacheKey()
	c.serverCapabilitiesMu.RLock()
	if cached, ok := c.serverCapabilities[cacheKey]; ok {
		c.serverCapabilitiesMu.RUnlock()
		return cloneDiscoInfo(cached), nil
	}
	c.serverCapabilitiesMu.RUnlock()

	iq, err := stanza.NewIQ(stanza.Attrs{Type: stanza.IQTypeGet, To: c.config.Domain})
	if err != nil {
		return nil, err
	}
	iq.DiscoInfo()

	resultCh, err := c.SendIQ(ctx, iq)
	if err != nil {
		return nil, err
	}

	select {
	case res, ok := <-resultCh:
		if !ok {
			return nil, errors.New("server capabilities request was cancelled")
		}
		payload, ok := res.Payload.(*stanza.DiscoInfo)
		if !ok || payload == nil {
			return nil, errors.New("server capabilities response did not contain disco info")
		}

		c.serverCapabilitiesMu.Lock()
		c.serverCapabilities[cacheKey] = cloneDiscoInfo(payload)
		c.serverCapabilitiesMu.Unlock()

		return cloneDiscoInfo(payload), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return nil, errors.New("server capabilities request failed")
}

func (c *Client) serverCapabilitiesCacheKey() string {
	if c.Session != nil {
		caps := c.Session.ServerFeatures.Caps
		if caps.Node != "" || caps.Ver != "" {
			return caps.Node + "#" + caps.Ver
		}

		caps = c.Session.Features.Caps
		if caps.Node != "" || caps.Ver != "" {
			return caps.Node + "#" + caps.Ver
		}
	}

	return c.config.Domain
}

func cloneDiscoInfo(info *stanza.DiscoInfo) *stanza.DiscoInfo {
	if info == nil {
		return nil
	}

	clone := *info
	clone.Identity = append([]stanza.Identity(nil), info.Identity...)
	clone.Features = append([]stanza.Feature(nil), info.Features...)
	if info.ResultSet != nil {
		resultSet := *info.ResultSet
		clone.ResultSet = &resultSet
	}

	return &clone
}

// Send marshals XMPP stanza and sends it to the server.
func (c *Client) Send(packet stanza.Packet) error {
	conn := c.transport
	if conn == nil {
		return errors.New("client is not connected")
	}
	if c.Session == nil {
		return errors.New("client session is not established")
	}

	data, err := xml.Marshal(packet)
	if err != nil {
		return errors.New("cannot marshal packet " + err.Error())
	}

	// Store stanza as non-acked as part of stream management
	// See https://xmpp.org/extensions/xep-0198.html#scenarios
	if c.config.StreamManagementEnable {
		if _, ok := packet.(stanza.SMRequest); !ok {
			if _, ok := packet.(stanza.SMAnswer); ok {
				return c.sendWithWriter(c.transport, data)
			}
			if c.Session.SMState.UnAckQueue == nil {
				return errors.New("stream management queue is not initialized")
			}
			toStore := stanza.UnAckedStz{Stz: string(data)}
			c.Session.SMState.UnAckQueue.Push(&toStore)
		}
	}

	return c.sendWithWriter(c.transport, data)
}

// SendIQ sends an IQ set or get stanza to the server. If a result is received
// the provided handler function will automatically be called.
//
// The provided context should have a timeout to prevent the client from waiting
// forever for an IQ result. For example:
//
//	ctx, _ := context.WithTimeout(context.Background(), 30 * time.Second)
//	result := <- client.SendIQ(ctx, iq)
func (c *Client) SendIQ(ctx context.Context, iq *stanza.IQ) (chan stanza.IQ, error) {
	if iq.Attrs.Type != stanza.IQTypeSet && iq.Attrs.Type != stanza.IQTypeGet {
		return nil, ErrCanOnlySendGetOrSetIq
	}
	resultCh := c.router.NewIQResultRoute(ctx, iq.Attrs.Id)
	if err := c.Send(iq); err != nil {
		c.router.RemoveIQResultRoute(iq.Attrs.Id)
		return nil, err
	}
	return resultCh, nil
}

// SendRaw sends an XMPP stanza as a string to the server.
// It can be invalid XML or XMPP content. In that case, the server will
// disconnect the client. It is up to the user of this method to
// carefully craft the XML content to produce valid XMPP.
func (c *Client) SendRaw(packet string) error {
	conn := c.transport
	if conn == nil {
		return errors.New("client is not connected")
	}
	if c.Session == nil {
		return errors.New("client session is not established")
	}

	// Store stanza as non-acked as part of stream management
	// See https://xmpp.org/extensions/xep-0198.html#scenarios
	if c.config.StreamManagementEnable {
		if c.Session.SMState.UnAckQueue == nil {
			return errors.New("stream management queue is not initialized")
		}
		toStore := stanza.UnAckedStz{Stz: packet}
		c.Session.SMState.UnAckQueue.Push(&toStore)
	}
	return c.sendWithWriter(c.transport, []byte(packet))
}

func (c *Client) sendWithWriter(writer io.Writer, packet []byte) error {
	var err error
	_, err = writer.Write(packet)
	return err
}

// ============================================================================
// Go routines

// Loop: Receive data from server
func (c *Client) recv(keepaliveQuit chan<- struct{}) {
	defer close(keepaliveQuit)

	for {
		val, err := stanza.NextPacket(c.transport.GetDecoder())
		if err != nil {
			c.ErrorHandler(err)
			c.disconnected(c.Session.SMState)
			return
		}

		// Handle stream errors
		switch packet := val.(type) {
		case stanza.StreamError:
			c.router.route(c, val)
			if packet.Error.Local == "see-other-host" && packet.SeeOtherHost != "" && c.Session != nil {
				c.Session.SMState.preferredReconAddr = packet.SeeOtherHost
			}
			c.streamError(packet.Error.Local, packet.Text)
			c.ErrorHandler(errors.New("stream error: " + packet.Error.Local))
			// We don't return here, because we want to wait for the stream close tag from the server, or timeout.
			c.Disconnect()
		// Process Stream management nonzas
		case stanza.SMRequest:
			answer := stanza.SMAnswer{XMLName: xml.Name{
				Space: stanza.NSStreamManagement,
				Local: "a",
			}, H: c.Session.SMState.Inbound}
			err = c.Send(answer)
			if err != nil {
				c.ErrorHandler(err)
				return
			}
			continue
		case stanza.SMAnswer:
			c.router.route(c, val)
			continue
		case stanza.StreamClosePacket:
			// TCP messages should arrive in order, so we can expect to get nothing more after this occurs
			c.transport.ReceivedStreamClose()
			c.Disconnect()
			continue
		case stanza.Message, stanza.Presence, *stanza.IQ:
			c.Session.SMState.Inbound++
		}
		// Preserve stanza order for callers that rely on arrival order.
		c.router.route(c, val)
	}
}

// Loop: send whitespace keepalive to server
// This is use to keep the connection open, but also to detect connection loss
// and trigger proper client connection shutdown.
func keepalive(transport Transport, interval time.Duration, quit <-chan struct{}) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			if err := transport.Ping(); err != nil {
				// When keepalive fails, we force close the transport. In all cases, the recv will also fail.
				ticker.Stop()
				_ = transport.Close()
				return
			}
		case <-quit:
			ticker.Stop()
			return
		}
	}
}
