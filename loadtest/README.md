Load Test for Khan API
======================

This application performs a specified amount of sequential requests to a remote Khan API server, with a specified time period between two consecutive requests. Usage: `../khan loadtest --help`

# Request parameters
Both amount and period are defined under the key `loadtest` within `../config/local.yaml`:
```
loadtest:
  requests:
    amount: 1
    period:
      ms: 1
```
Or setting the following environment variables:
```
KHAN_LOADTEST_REQUESTS_AMOUNT (default: 1)
KHAN_LOADTEST_REQUESTS_PERIOD_MS (default: 1)
```

# Request operations
For each request, an operation is chosen at random with probabilities defined at `../config/local.yaml`, like this:
```
loadtest:
  operations:
    retrieveSharedClan:
      probability: 0.3
    updateSharedClan:
      probability: 0.3
```
Or setting the following environment variables:
```
KHAN_LOADTEST_OPERATIONS_RETRIEVESHAREDCLAN_PROBABILITY (no default value)
KHAN_LOADTEST_OPERATIONS_UPDATESHAREDCLAN_PROBABILITY (no default value)
```
Request operations are defined in files `player.go`, `clan.go` and `membership.go`. Check the operation keys to set the probabilities.

# Remote KHAN API server parameters 
Khan API server URL and credentials should be set by environment variables, like this:
```
KHAN_KHAN_URL: URL including protocol and port to remote Khan API server
KHAN_KHAN_USER: basic auth username
KHAN_KHAN_PASS: basic auth password
KHAN_KHAN_GAMEID: game public ID
KHAN_KHAN_TIMEOUT: nanoseconds to wait before timing out a request (default: 500 ms)
KHAN_KHAN_MAXIDLECONNS: max keep-alive connections to keep among all hosts (default: 100)
KHAN_KHAN_MAXIDLECONNSPERHOST: max keep-alive connections to keep per-host (default: 2)
```

# Operations with clans shared among different load test processes
Some operations are targeted to a set of clans that should be shared among different processes running a load test. This is supposed to test the most common sequence of operations (`updatePlayer`, then `getClan`, then `updateClan`) in its most common use case, which is a number of different clients updating the score of a particular clan within its metadata field. Use the file `../config/loadTestSharedClans.yaml` to specify a list of clan public IDs, like this:

```
clans:
- "clan1publicID"
- "clan2publicID"
```
