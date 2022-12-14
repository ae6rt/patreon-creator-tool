Simple Patreon tool based on the Patreon V2 API that fetches creator campaign and pledge info.  The tool is by no means
comprehensive but it provides something to build on.

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

## References

* [Patreon API](https://docs.patreon.com/#introduction)
* [Register Patreon OAuth clients](https://www.patreon.com/portal/registration/register-clients)
