# Issue #109: Code clean-up

**Opened by** mremond **at** 2019-09-06T07:39:57Z

## Body

Client manager should only call Resume() when the client support Stream management. It should not call resume for Components, but directly call connect.
The clean-up should make the component code clearer, avoid to make all the connection process in resume.

See #108 
