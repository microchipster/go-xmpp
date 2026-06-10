# PR #29: WIP: use xml Name

**Opened by** genofire **at** 2019-06-05T01:38:19Z

## Body



## Review comments (on diffs)

---
**genofire** at 2019-06-05T01:40:10Z on registry.go (pos 17)
Does not work alike expected yet.
We need to read the golang xml tag annotation -> a new created `xml.Name` does not contain a value in `Local` and `Space`

## Reviews

---
**genofire** at 2019-06-05T01:40:16Z — COMMENTED

