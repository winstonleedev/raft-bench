// Copyright 2017,2018 Lei Ni (nilei81@gmail.com).
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
ondisk is an example program for dragonboat's on disk state machine.
*/

package dragonboat

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/lni/dragonboat/v4"
	"github.com/lni/dragonboat/v4/config"
	"github.com/lni/dragonboat/v4/logger"

	"github.com/thanhphu/raftbench/util"
)

type RequestType uint64

const (
	exampleClusterID uint64 = 128
)

func Main(cluster string, nodeID int, addr string, join bool, test util.TestParams) {
	if len(addr) == 0 && nodeID != 1 && nodeID != 2 && nodeID != 3 {
		fmt.Fprintf(os.Stderr, "node id must be 1, 2 or 3 when address is not specified\n")
		os.Exit(1)
	}
	// https://github.com/golang/go/issues/17393
	if runtime.GOOS == "darwin" {
		signal.Ignore(syscall.Signal(0xd))
	}
	initialMembers := make(map[uint64]string)
	if !join {
		addresses := strings.Split(strings.Replace(cluster, "http://", "", -1), ",")
		for idx, v := range addresses {
			initialMembers[uint64(idx+1)] = v
		}
	}
	var nodeAddr string
	if len(addr) != 0 {
		nodeAddr = addr
	} else {
		nodeAddr = initialMembers[uint64(nodeID)]
	}
	fmt.Fprintf(os.Stdout, "node address: %s\n", nodeAddr)
	logger.GetLogger("raft").SetLevel(logger.ERROR)
	logger.GetLogger("rsm").SetLevel(logger.WARNING)
	logger.GetLogger("transport").SetLevel(logger.WARNING)
	logger.GetLogger("grpc").SetLevel(logger.WARNING)
	rc := config.Config{
		ReplicaID:              uint64(nodeID),
		ShardID:                exampleClusterID,
		ElectionRTT:            10,
		HeartbeatRTT:           1,
		CheckQuorum:            true,
		SnapshotEntries:        0,
		DisableAutoCompactions: true,
		MaxInMemLogSize:        0,
		CompactionOverhead:     5,
	}
	datadir := filepath.Join(
		fmt.Sprintf("wal-dragonboat-%d", nodeID))
	nhc := config.NodeHostConfig{
		WALDir:         datadir,
		NodeHostDir:    datadir,
		RTTMillisecond: 200,
		RaftAddress:    nodeAddr,
	}
	nh, err := dragonboat.NewNodeHost(nhc)
	if err != nil {
		panic(err)
	}
	if err := nh.StartReplica(initialMembers, join, NewMemKV, rc); err != nil {
		fmt.Fprintf(os.Stderr, "failed to add cluster, %v\n", err)
		os.Exit(1)
	}

	cs := nh.GetNoOPSession(exampleClusterID)

	// Wait for shard to become ready.
	time.Sleep(2 * time.Second)

	ctx, _ := context.WithTimeout(context.Background(), 1000*time.Second)

	util.Bench(test, func(k string) bool {
		_, err := nh.SyncRead(ctx, exampleClusterID, k)
		return err == nil
	}, func(k string, v string) bool {
		kv := &KVData{
			Key: k,
			Val: v,
		}
		data, err := json.Marshal(kv)
		if err != nil {
			return false
		}

		_, err = nh.SyncPropose(ctx, cs, data)
		return err == nil
	})
}
