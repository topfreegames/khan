Khan's Benchmarks
=================

You can see khan's benchmarks in our [CI server](https://travis-ci.org/topfreegames/khan/) as they get run with every build.

## Creating the performance database

To create a new database for running your benchmarks, just run:

```
$ make drop-perf migrate-perf
```

## Running Benchmarks

If you want to run your own benchmarks, just download the project, and run:

```
$ make run-test-khan run-perf
```

## Generating test data

If you want to run your perf tests against a database with more volume of data, just run this command prior to running the above one:

```
$ make drop-perf migrate-perf db-perf
```

**Warning**: This will take a long time running (around 30m).

## Results

The results should be similar to these:

```
BenchmarkCreateClan-2                  	    2000	   3053999 ns/op
BenchmarkUpdateClan-2                  	    2000	   2000650 ns/op
BenchmarkRetrieveClan-2                	     500	  10522248 ns/op
BenchmarkRetrieveClanSummary-2         	    5000	   1187486 ns/op
BenchmarkSearchClan-2                  	    5000	   1205325 ns/op
BenchmarkListClans-2                   	    5000	   1135555 ns/op
BenchmarkLeaveClan-2                   	    1000	   3824284 ns/op
BenchmarkTransferOwnership-2           	     500	   8642818 ns/op
BenchmarkCreateGame-2                  	    3000	   1248042 ns/op
BenchmarkUpdateGame-2                  	    2000	   2141705 ns/op
BenchmarkApplyForMembership-2          	    1000	   5695344 ns/op
BenchmarkInviteForMembership-2         	     500	   8916792 ns/op
BenchmarkApproveMembershipApplication-2	     500	  13480574 ns/op
BenchmarkApproveMembershipInvitation-2 	    1000	  10517905 ns/op
BenchmarkDeleteMembership-2            	     500	   9548314 ns/op
BenchmarkPromoteMembership-2           	     500	   8961424 ns/op
BenchmarkDemoteMembership-2            	     500	   9202060 ns/op
BenchmarkCreatePlayer-2                	    3000	   1344267 ns/op
BenchmarkUpdatePlayer-2                	    3000	   1829329 ns/op
BenchmarkRetrievePlayer-2              	     300	  14412830 ns/op
```
