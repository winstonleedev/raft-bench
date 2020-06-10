#!/usr/bin/env bash

# exit when any command fails
set -e

for s in raft0 raft1 raft2
do
  ssh -o "StrictHostKeyChecking no" ubuntu@${s} "cd ~/raft-bench && rm -rf wal-* && killall -9 raftbench && git pull && /usr/local/go/bin/go build"
done

ssh ubuntu@raft0 "cd ~/raft-bench && ./raftbench --engine hashi --inmem --nodeid node1 --haddr raft0:11000 --raddr raft0:12000 ~/raft-bench/wal-hashi" &
ssh ubuntu@raft1 "cd ~/raft-bench && ./raftbench --engine hashi --inmem --nodeid node2 --haddr raft1:11000 --raddr raft1:12000 --joinaddr raft0:11000 ~/raft-bench/wal-hashi" &
ssh ubuntu@raft2 "cd ~/raft-bench && ./raftbench --engine hashi --inmem --nodeid node3 --haddr raft2:11000 --raddr raft2:12000 --joinaddr raft0:11000 ~/raft-bench/wal-hashi"

