// Copyright 2015 The etcd Authors
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

package etcd

import (
	"strings"

	"github.com/coreos/etcd/raft/raftpb"

	"github.com/thanhphu/raftbench/util"
)

// Main function
func Main(cluster string, id int, _ int, join bool, test util.TestParams) {
	proposeC := make(chan string)
	defer close(proposeC)
	confChangeC := make(chan raftpb.ConfChange)
	defer close(confChangeC)

	// raft provides a commit stream for the proposals from the http api
	var kvs *keystore
	getSnapshot := func() ([]byte, error) { return kvs.getSnapshot() }
	commitC, errorC, snapshotterReady := newRaftNode(id, strings.Split(cluster, ","), join, getSnapshot, proposeC, confChangeC)

	kvs = newKVStore(<-snapshotterReady, proposeC, commitC, errorC)

	util.Bench(test, func(k string) bool {
		kvs.Lookup(k)
		return true
	}, func(k string, v string) bool {
		kvs.Propose(k, v)
		return true
	})
}
