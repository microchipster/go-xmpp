# Issue #44: Improve error formatting / wrapping / unwrapping

**Opened by** mremond **at** 2019-06-07T13:39:58Z

## Body



## Comments

---
**wichert** at 2019-12-09T13:29:46Z
Is this about getting rid of xerrors?

---
**mremond** at 2019-12-09T14:11:49Z
It was mostly about moving to the new Go error handling from Go 1.13. This means moving from xerrors to new standard code, but also be more consistent in how we report the error and wrap some context.
