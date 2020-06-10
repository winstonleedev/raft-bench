#!/usr/bin/env bash

# exit when any command fails
set -e

go build

scp raftbench ubuntu@raft0:~/raft-bench &
scp raftbench ubuntu@raft1:~/raft-bench &
scp raftbench ubuntu@raft2:~/raft-bench

ssh ubuntu@raft0 "rm -rf ~/raft-bench/wal-*"
ssh ubuntu@raft1 "rm -rf ~/raft-bench/wal-*"
ssh ubuntu@raft2 "rm -rf ~/raft-bench/wal-*"

ssh ubuntu@raft0 "cd ~/raft-bench && ./raftbench --engine $1 --id 1 --cluster http://10.0.12.142:12379,http://10.0.12.174:12379,http://10.0.12.37:12379  --test" &
ssh ubuntu@raft1 "cd ~/raft-bench && ./raftbench --engine $1 --id 2 --cluster http://10.0.12.142:12379,http://10.0.12.174:12379,http://10.0.12.37:12379 " &
ssh ubuntu@raft2 "cd ~/raft-bench && ./raftbench --engine $1 --id 3 --cluster http://10.0.12.142:12379,http://10.0.12.174:12379,http://10.0.12.37:12379 "

