#!/usr/bin/env bash

# exit when any command fails
set -e

go build
scp raftbench ubuntu@raft0:~/raft-bench
scp raftbench ubuntu@raft1:~/raft-bench
scp raftbench ubuntu@raft2:~/raft-bench
ssh ubuntu@raft0 'rm -rf ~/raft-bench/wal-*'
ssh ubuntu@raft1 'rm -rf ~/raft-bench/wal-*'
ssh ubuntu@raft2 'rm -rf ~/raft-bench/wal-*'
ssh ubuntu@raft0 'cd ~/raft-bench && ./raftbench --engine etcd --id 1 --cluster http://raft0:12379,http://raft1:22379,http://raft2:32379 --test' &
ssh ubuntu@raft1 'cd ~/raft-bench && ./raftbench --engine etcd --id 2 --cluster http://raft0:12379,http://raft1:22379,http://raft2:32379' &
ssh ubuntu@raft2 'cd ~/raft-bench && ./raftbench --engine etcd --id 3 --cluster http://raft0:12379,http://raft1:22379,http://raft2:32379'