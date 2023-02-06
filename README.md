# raft-bench

raft-bench is a benchmark to compare the performance between
[etcd](../../raft), [hashicorp](../../raft) and [dragonboat](../../raft)'s RAFT library

It provides raw performance insight of a key-value store cluster backed by the [Raft][raft] consensus algorithm.

Benchmark results are as follow



[raft]: http://raftconsensus.github.io/

## Getting Started

### Install prequisites

```bash
sudo apt install -y nano librocksdb-dev golang psmisc
```


### Building raft-bench

Clone `raft-bench` 

```sh
git clone https://github.com/thanhphu/raft-bench.git
cd raft-bench
go build
```

Use `raft-bench --help` for a list of parameters

### Running a local cluster

First install [goreman](https://github.com/mattn/goreman), which manages Procfile-based applications.

The [Procfile script](https://github.com/thanhphu/raft-bench/blob/master/Procfile) will set up a local example cluster. Start it with:

```sh
goreman -f <procfile> start
```

__procfile__ can be

* `Procfile` for etcd
* `Procfile-hashi` for hashicorp
* `Procfile-dragonboat` for dragonboat

Example: `Procfile` for etcd

```bash
# Use goreman to run `go get github.com/mattn/goreman`err <-
raftbench1: ./raftbench --engine etcd --id 1 --cluster http://127.0.0.1:12379,http://127.0.0.1:22379,http://127.0.0.1:32379 --port 12380 --test
raftbench2: ./raftbench --engine etcd --id 2 --cluster http://127.0.0.1:12379,http://127.0.0.1:22379,http://127.0.0.1:32379 --port 22380
raftbench3: ./raftbench --engine etcd --id 3 --cluster http://127.0.0.1:12379,http://127.0.0.1:22379,http://127.0.0.1:32379 --port 32380
```

This will bring up three raft-bench instances.

The instance with the `--test` parameter will perform write benchmark to itself and distribute the state to other instances

### Running a remote cluster

Set up 3 machines with the prequisites, named raft0, raft1, raft2 reachable from development machine
and each other. Check out raft-bench to `~/raft-bench`

On development machine, run one of the 3

```sh
./start-etcd.sh
./start-hashi.sh
./start-dragonboat.sh
```

It will pull the latest source code on the remote machine, build it and start the benchmark

### Result

The program write results to the csv pointed out by `--logfile` parameter. A sample result looks like this

```
write,1,1000,1000,6264754
read,1,1000,1000,1854152
write,2,1000,1000,6021356
read,2,1000,1000,1827525
write,3,1000,1000,5960166
read,3,1000,1000,1841903
write,4,1000,1000,5964795
read,4,1821,1000,2462557
write,5,3000,1000,3435424
read,5,3000,1000,3416640
write,6,3000,1000,3435198
read,6,3000,1000,3408357
write,7,3000,1000,3435068
read,7,3000,1000,3422464
write,8,3000,1000,3427716
read,8,3000,1000,3409562
write,9,3000,1000,3426177
read,9,3000,1000,3413216
write,10,3000,1000,3430132
read,10,3000,1000,3413698
```

## Design

The raft-bench consists of three components: a raft-backed key-value store, a REST API server, 
and a raft consensus server based on etcd's raft implementation.

The raft-backed key-value store is a key-value map that holds all committed key-values.
The store bridges communication between the raft server and the REST server.
Key-value updates are issued through the store to the raft server.
The store updates its map once raft reports the updates are committed.

The REST server exposes the current raft consensus by accessing the raft-backed key-value store.
A GET command looks up a key in the store and returns the value, if any.
A key-value PUT command issues an update proposal to the store.

The raft server participates in consensus with its cluster peers.
When the REST server submits a proposal, the raft server transmits the proposal to its peers.
When raft reaches a consensus, the server publishes all committed updates over a commit channel.
For raft-bench, this commit channel is consumed by the key-value store.

