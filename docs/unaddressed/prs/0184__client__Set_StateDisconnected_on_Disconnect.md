# PR #184: [client] Set StateDisconnected on Disconnect

**Opened by** jaredledvina **at** 2025-04-12T14:48:23Z

## Body

This PR is a small update so that when the client's `Disconnect()` function is called, we also set the current state to disconnected. Currently, subsequent `connect` fail with https://github.com/FluuxIO/go-xmpp/blob/7186c058fd8eb4985efa52ed00ddbef30371c211/stream_manager.go#L131 as we never transition the state to disconnected. 
