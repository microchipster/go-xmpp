# Issue #122: Check if the TransportConfiguration struct level is really needed in the config

**Opened by** mremond **at** 2019-10-29T13:41:25Z

## Body

As the transport is encode in the url scheme, maybe we do not need that extra level and we can simply have a flat structure configuration.
