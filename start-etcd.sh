#!/usr/bin/env bash

for s in raft0 raft1 raft2
do
  ssh -o "StrictHostKeyChecking no" ubuntu@${s} "cd ~/raft-bench && rm -rf wal-* && killall -9 raftbench && git pull && /usr/local/go/bin/go build"
done

# exit when any command fails
set -e

for s in raft0 raft1 raft2
do
  ssh -o "StrictHostKeyChecking no" ubuntu@${s} "cd ~/raft-bench && /usr/local/go/bin/go build"
done

ssh ubuntu@raft0 "cd ~/raft-bench && ./raftbench --engine etcd --mil 1000 --firstWait 5000 --step 1 --id 1 --cluster http://raft0:12379,http://raft1:12379,http://raft2:12379 --test --logfile etcd.csv" &
ssh ubuntu@raft1 "cd ~/raft-bench && ./raftbench --engine etcd --id 2 --cluster http://raft0:12379,http://raft1:12379,http://raft2:12379 " &
ssh ubuntu@raft2 "cd ~/raft-bench && ./raftbench --engine etcd --id 3 --cluster http://raft0:12379,http://raft1:12379,http://raft2:12379 "

