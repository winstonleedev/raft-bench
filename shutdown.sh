#!/usr/bin/env bash

set +e

for s in raft0 raft1 raft2
do
  ssh -o "StrictHostKeyChecking no" ubuntu@${s} "cd ~/raft-bench && rm -rf wal-* && killall -9 raftbench && git pull && /usr/local/go/bin/go build"
done
