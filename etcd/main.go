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
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/coreos/etcd/raft/raftpb"
)

const (
	numKeys = 1
	mil     = 1000000
)

// Main function
func Main(cluster string, id int, _ int, join bool, test bool) {
	proposeC := make(chan string)
	defer close(proposeC)
	confChangeC := make(chan raftpb.ConfChange)
	defer close(confChangeC)

	// raft provides a commit stream for the proposals from the http api
	var kvs *keystore
	getSnapshot := func() ([]byte, error) { return kvs.getSnapshot() }
	commitC, errorC, snapshotterReady := newRaftNode(id, strings.Split(cluster, ","), join, getSnapshot, proposeC, confChangeC)

	kvs = newKVStore(<-snapshotterReady, proposeC, commitC, errorC)

	if test {
		for i := 0; i < 3; i++ {
			time.Sleep(3000)

			start := time.Now()
			k := 0
			for k < numKeys*mil {
				v := rand.Int()
				go kvs.Propose(string(k), string(v))
				k += 1
			}
			fmt.Printf("Write test, %v, %v, %v\n", i+1, numKeys*mil, time.Since(start))

			time.Sleep(3000)
			start = time.Now()
			k = 0
			for k < numKeys*mil {
				go kvs.Lookup(string(k))
				k += 1
			}
			fmt.Printf("Read test, %v, %v, %v\n", i+1, numKeys*mil, time.Since(start))
		}
	}
}
