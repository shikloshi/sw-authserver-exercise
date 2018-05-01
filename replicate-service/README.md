# replicate-service

This process is used for replicating an http request and sending it to multiple http server, it will return the first response returned by one of the backends back to the client.

Config
=================
* `BACKENDS` - comma separated list of addresses (`host:port`) to use as backends.
* `MAX_RETRIES` - number of send retries after failed sending to backend.
* `FAILED_REQUESTS_QUEUE_SIZE` - buffer to hold failed requests for background retry.
* `RETRY_WORKERS_COUNT` - how many go routines should handle the requests retry queue.
* `INIT_RETRY_WAIT` - initial wait for retry with exponential backoff - time in ms.

Metrics
=============
Exposed via the go `expvar` package (can be "easily" instrumented with data-dog or prometheus):

1. retries - number of retries
2. incoming\_requests - number of incoming requests
3. first\_backend\_response_time - time to get first response in nanoseconds

All metrics are exposes under `/debug/vars`


Road map :)
===================

- [ ] smart manage of http clients - pool of persistant connection clients to backends.
- [ ] backend healthchecks.
- [ ] clean exit - signal all channels and flush all waiting jobs.
- [ ] bug: wrong response when all backend failed to return ok response on first try.
- [ ] command line flags using `flag` package (with env var default fallback).