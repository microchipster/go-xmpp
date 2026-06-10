# PR #180: Fix passing Component to StreamManager

**Opened by** bodqhrohro **at** 2021-12-18T15:57:51Z

## Body

0a4acd12c34b0048ff7554b7c56ddce975eb9286, which fixes #160, introduced
a regression as it assumed only Client may implement StreamClient, and
passing a component triggers an unconditional "client is not
disconnected" error.

## Reviews

---
**initpwn** at 2022-04-20T11:00:09Z — APPROVED

