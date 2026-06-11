# Issue #86: Cert checker should be able to check both starttls and tls cert

**Opened by** mremond **at** 2019-07-09T16:14:40Z

## Body


---
**Addressed note**

The cert checker now supports both XMPP STARTTLS and direct TLS certificate validation. `ServerCheck.Check()` splits into STARTTLS and direct-TLS paths, auto-detects direct TLS for port `5223`, and validates the presented server certificate chain against the system CA pool with hostname and expiration checks. Regression tests cover both modes.

---
