# PR #152: Add Server-To-Client handler

**Opened by** vduduh **at** 2020-02-28T07:22:43Z

## Body

https://xmpp.org/extensions/xep-0199.html#s2c

## Issue comments

---
**p1bot** at 2020-02-28T07:22:45Z
Hi @vduduh, many thanks for your contribution!

In order for us to evaluate and accept your PR, we ask that you [sign a contribution license agreement](https://cla.process-one.net/). It's all electronic and will take just minutes.


---
**p1bot** at 2020-02-28T07:26:30Z
You did it @vduduh!

Thank you for signing the ProcessOne Contribution License Agreement.

We will have a look at your contribution!


---
**remicorniere** at 2020-02-28T09:19:59Z
Hello vduduh and thanks for contributing !
As pings described in XEP-0199 are IQ stanzas, and not nonzas, I'd rather have them processed by the router instead, in this function :   

`
func (r *Router) route(s Sender, p stanza.Packet) 
`

Also, as the spec that you linked says, the client must only respond if it supports the ping namespace. For now, there is no way for the client to say that it supports this feature because responding to IQ requests is not supported (also, we don't support it). We would need to implement that first.  

If you'd like to have a go at it, that'd be great ! :)  


---
**vduduh** at 2020-02-28T12:08:36Z
Hello @remicorniere !
Some servers force disconnection of clients if they do not respond to ping. That is why I implemented this functionality.

---
**remicorniere** at 2020-02-28T14:05:47Z
I understand. Thank you for suggesting this change.   

Although for now I believe the code edits were not made in the right place (see my previous comment).  
You can edit your patch to implement the suggested adjustments. Please let me know if you have any questions or need guidance ! :)  
Also please refer to the [guidelines](https://github.com/FluuxIO/go-xmpp/blob/master/CONTRIBUTING.md): we are missing tests for your new feature. Could you please add some ?

For now I will link this PR to an issue reflecting your feature request.  
Thanks ! 
