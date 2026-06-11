# Issue #101: MUC Implementation

**Opened by** kylegibbons **at** 2019-08-13T18:54:10Z

## Body

Is joining a MUC implemented in this library?

I looked around and found a few issues mentioning MUC that seem to indicate it is implemented, but I can't find any documentation on how to actually use it.

Could you point me in the right direction? Thanks!

## Comments

---
**Kroev** at 2019-08-26T09:44:14Z
Yeah i had to fiddle a bit too but joining a MUC is simple, once you got the hang of it. Looking into XMPP documentation also helps, see [XEP0045](https://xmpp.org/extensions/xep-0045.html#enter).

I did it in the postconnect handler of the streammanger here but you should be able to do it whenever you want, as long as the client is connected 
```golang
func postConnect(s xmpp.Sender) {
	client, ok := s.(*xmpp.Client)
	if !ok {
		fmt.Println("post connect sender not a client, cannot proceed")
		//error handling 
	}
       
        //I used uuids to make sure the id is unique
	uuid, _ := uuid.NewV4()
	id := uuid.String()

	//prepare presence for joining the MUC
	presence := stanza.NewPresence(stanza.Attrs{
		To:   "<mucroom>@<xmppdomain_supporting_muc>/<muc_nickname>",//fill the fields accordingly
		From: client.Session.BindJid,
		Id:   id,
	})
        
        //as stated in the XEP0045 documentation, you have to tell that you are able to speak muc
	presence.Extensions = append(presence.Extensions, stanza.MucPresence{})

        //send the stuff and actually join the MUC
	err := client.Send(presence)
	if err != nil {
		//err handling
	}
}
```

---
**Addressed note**

The library already includes MUC stanza support and tests, and the issue is now treated as implemented/usage guidance rather than missing core functionality.
