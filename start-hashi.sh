#!/usr/bin/env bash

# exit when any command fails
set -e

go build

scp raftbench ubuntu@raft0:~/raft-bench &
scp raftbench ubuntu@raft1:~/raft-bench &
scp raftbench ubuntu@raft2:~/raft-bench

ssh ubuntu@raft0 "rm -rf ~/raft-bench/wal-*" &
ssh ubuntu@raft1 "rm -rf ~/raft-bench/wal-*" &
ssh ubuntu@raft2 "rm -rf ~/raft-bench/wal-*"

ssh ubuntu@raft0 "cd ~/raft-bench && ./raftbench --engine hashi --inmem --nodeid node1 --haddr :11000 --raddr :12000 --test ./wal-hashi-1" &
ssh ubuntu@raft1 "cd ~/raft-bench && ./raftbench --engine hashi --inmem --nodeid node2 --haddr :11000 --raddr :12000 --joinaddr raft0:11000 ./wal-hashi-2" &
ssh ubuntu@raft2 "cd ~/raft-bench && ./raftbench --engine hashi --inmem --nodeid node3 --haddr :11000 --raddr :12000 --joinaddr raft0:11000 ./wal-hashi-3"

./raftbench --engine hashi --inmem --nodeid node1 --haddr :11000 --raddr :12000 --test ./wal-hashi-1
./raftbench --engine hashi --inmem --nodeid node2 --haddr :11000 --raddr :12000 --joinaddr raft0:11000 ./wal-hashi-2
./raftbench --engine hashi --inmem --nodeid node3 --haddr :11000 --raddr :12000 --joinaddr raft0:11000 ./wal-hashi-3