#!/usr/bin/env bash

# exit when any command fails
set -e

ssh ubuntu@raft0 "rm -rf ~/raft-bench/wal-*"
ssh ubuntu@raft1 "rm -rf ~/raft-bench/wal-*"
ssh ubuntu@raft2 "rm -rf ~/raft-bench/wal-*"

ssh ubuntu@raft0 "cd ~/raft-bench && git pull && /usr/local/go/bin/go build && ./raftbench --engine etcd --id 1 --cluster http://raft0:12379,http://raft1:12379,http://raft2:12379  --test" &
ssh ubuntu@raft1 "cd ~/raft-bench && git pull && /usr/local/go/bin/go build && ./raftbench --engine etcd --id 2 --cluster http://raft0:12379,http://raft1:12379,http://raft2:12379 " &
ssh ubuntu@raft2 "cd ~/raft-bench && git pull && /usr/local/go/bin/go build && ./raftbench --engine etcd --id 3 --cluster http://raft0:12379,http://raft1:12379,http://raft2:12379 "

