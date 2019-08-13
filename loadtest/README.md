Load Test for Khan API
======================

This application performs a random sequence of a specified amount of operations on a remote Khan API server, with a specified time period between two consecutive operations. It also allows multiple goroutines for local concurrency (multiple concurrent random sequences). Usage: `../khan loadtest --help`

# Game parameters
Khan does not offer a route to get game information, so the membership level for application and the maximum number of members per clan are defined under the key `loadtest.game` within `../config/local.yaml`:
```
loadtest:
  game:
    membershipLevel: "member"
    maxMembers: 50
```
Or setting the following environment variables:
```
KHAN_LOADTEST_GAME_MEMBERSHIPLEVEL (default: "")
KHAN_LOADTEST_GAME_MAXMEMBERS (default: 0)
```

# Client parameters 
Client parameters are defined under the key `loadtest.client` within `../config/local.yaml`:
```
loadtest:
  client:
    url: "http://localhost:8080"
    gameid: "epiccardgame"
```
Or setting the following environment variables:
```
KHAN_LOADTEST_CLIENT_URL: URL including protocol (http/https) and port to remote Khan API server (default: "")
KHAN_LOADTEST_CLIENT_USER: basic auth username (default: "")
KHAN_LOADTEST_CLIENT_PASS: basic auth password (default: "")
KHAN_LOADTEST_CLIENT_GAMEID: game public ID (default: "")
KHAN_LOADTEST_CLIENT_TIMEOUT: nanoseconds to wait before timing out a request (default: 500 ms)
KHAN_LOADTEST_CLIENT_MAXIDLECONNS: max keep-alive connections to keep among all hosts (default: 100)
KHAN_LOADTEST_CLIENT_MAXIDLECONNSPERHOST: max keep-alive connections to keep per-host (default: 2)
```

# Operation parameters
The amount of operations per sequence/goroutine, the time period between two consecutive operations and the probabilities per operation are defined under the key `loadtest.operations` within `../config/local.yaml`:
```
loadtest:
  operations:
    amount: 1
    period:
      ms: 1
    updateSharedClanScore:
      probability: 0.8
    createPlayer:
      probability: 0.01
    createClan:
      probability: 0.01
    leaveClan:
      probability: 0.01
    transferClanOwnership:
      probability: 0.01
    applyForMembership:
      probability: 0.01
    selfDeleteMembership:
      probability: 0.01
```
Or setting the following environment variables:
```
KHAN_LOADTEST_OPERATIONS_AMOUNT (default: 0)
KHAN_LOADTEST_OPERATIONS_PERIOD_MS (default: 0)
KHAN_LOADTEST_OPERATIONS_UPDATESHAREDCLANSCORE_PROBABILITY (default: 0.8)
KHAN_LOADTEST_OPERATIONS_CREATEPLAYER_PROBABILITY (default: 0.01)
KHAN_LOADTEST_OPERATIONS_CREATECLAN_PROBABILITY (default: 0.01)
KHAN_LOADTEST_OPERATIONS_LEAVECLAN_PROBABILITY (default: 0.01)
KHAN_LOADTEST_OPERATIONS_TRANSFERCLANOWNERSHIP_PROBABILITY (default: 0.01)
KHAN_LOADTEST_OPERATIONS_APPLYFORMEMBERSHIP_PROBABILITY (default: 0.01)
KHAN_LOADTEST_OPERATIONS_SELFDELETEMEMBERSHIP_PROBABILITY (default: 0.01)
```

# Operations with clans shared among different load test processes
Some operations are targeted to a set of clans that should be shared among different goroutines/sequences of operations. One of these operations is supposed to test the most common sequence of requests (`updatePlayer`, then `getClan`, then `updateClan`) in its most common use case, which is a number of different clients updating the score of a particular clan within its metadata field. Use the file `../config/loadTestSharedClans.yaml` to specify the list of public IDs for shared clans:

```
clans:
- "clan1publicID"
- "clan2publicID"
```
