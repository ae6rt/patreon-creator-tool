Simple Patreon tooling for fetching creator campaign and pledge info.  The tool is by no means
comprehensive, but it provides something to build on.

## Build

```
make linux
make mac
```

## Run 

```
./patreon-tool -get-pledges -access-token <the token>
```

which outputs simple member and tier records such as

```
fullName=Tom Bombadil email=tom@example.com pledgeAmount:500, tiers: The Works
```

Richer output is possible with code changes.
