# Issue #120: Lifecycle hooks are missing

**Opened by** wichert **at** 2019-10-27T08:42:59Z

## Body

I am missing the ability to hook into the stream manager and client lifecycle. Client exposes an event handler, but if you use StreamManager it already uses that, making it impossible to respond to state changes. For my purposes I need to respond to:

* authentication success, so I can fetch roster, vcards for contacts and a few other app-specific things
* authentication failure, so I can prompt for new credentials and retry auth
* successful reconnect, so I can update my local state


## Comments

---
**wichert** at 2019-10-28T14:04:55Z
I've been thinking a little bit about this. I don't think a channel is the right solution here, since it would block xmpp if nothing is listening on the channel. A callback/event approach like you currently have should be a better approach. If we can make this a bit more flexible to allow multiple subscribers and named event types. We can, for example, use https://github.com/kataras/go-events to allow something like this:

```go
client.On("status", func handle(payload …interface{}) {
    status := payload[0].(string)
    fmt.Printf("XMPP client changed state to %v\n", status)
})

client.On("status", func handle(payload …interface{}) {
    status := payload[0].(string)
    if status == "authenticated" {
        RetrieveRoster(client)
    }
}

client.On("reconnect", func handle(…interface{}) {
    fmt.Println("Stream manager is attempting to reconnect")
}
```

This can also replace `StreamManager.PostConnect`.


---
**wichert** at 2019-10-28T14:29:44Z
I just noticed StreamManager.PostConnect, which looks very much like a "successful (re)connect and auth" event. That might actually be enough for me at the moment.

---
**mremond** at 2019-11-04T08:55:13Z
@wichert Yes, I confirm that the role of postConnect is to perform the post auth operations.
However, I agree that adding a few more hooks may be nice. I will think about it.

---
**wichert** at 2019-11-04T08:58:28Z
I have already started to find some issues with postConnect: if for some reason the stream restarts, for example because some IQ processing error results in an error, the code I run in postConncect won't notice and happily keeps running. Perhaps one thing that might be useful is to have a per-stream context that is passed to callbacks, and that is cancelled when the stream ends for whatever reason.


---
**mremond** at 2019-11-04T09:00:14Z
Yes, adding the context for cancellation would be handy.
