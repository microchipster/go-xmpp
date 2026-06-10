# Issue #47: Add support for stream management

**Opened by** mremond **at** 2019-06-10T09:17:48Z

## Body



## Comments

---
**mremond** at 2019-07-31T16:48:24Z
Work-in-progress in this branch: https://github.com/FluuxIO/go-xmpp/tree/go-xmpp-47

---
**genofire** at 2019-09-05T19:25:52Z
Okay we have a big problem -> Component does not work:

- we should provide a better error handling (the Resume error is not handle in StreamManager) + revert example
**or**
- fix it for component soon as possible

---
**mremond** at 2019-09-05T19:51:30Z
Ok, I will have a look tomorrow. Thanks !
