Road map :)
===================

- [ ] smart manage of http clients - pool of persistant connection clients to backends.
- backend healthchecks.
- clean exit - signal all channels and flush all waiting jobs.
- bug: wrong response when all backend failed to return ok response on first try.
- command line flags using "flags" (with env var default fallback).