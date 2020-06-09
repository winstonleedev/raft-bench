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
	"bytes"
	"encoding/gob"
	"log"
	"sync"

	"github.com/coreos/etcd/snap"
)

// a key-value store backed by raft
type keystore struct {
	proposeC  chan<- string // channel for proposing updates
	mu        sync.RWMutex
	kvStore   map[string]string // current committed key-value pairs
	snapshots *snap.Snapshotter
}

type kv struct {
	Key string
	Val string
}

func newKVStore(snapshots *snap.Snapshotter, proposeC chan<- string, commitC <-chan *string, errorC <-chan error) *keystore {
	s := &keystore{
		proposeC:  proposeC,
		mu:        sync.RWMutex{},
		kvStore:   make(map[string]string),
		snapshots: snapshots,
	}
	// replay log into key-value map
	s.readCommits(commitC, errorC)
	// read commits from raft into kvStore map until error
	go s.readCommits(commitC, errorC)
	return s
}

func (s *keystore) Lookup(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.kvStore[key]
	return v, ok
}

func (s *keystore) Propose(k string, v string) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(kv{k, v}); err != nil {
		log.Fatal(err)
	}
	s.proposeC <- buf.String()
}

func (s *keystore) readCommits(commitC <-chan *string, errorC <-chan error) {
	for data := range commitC {
		if data == nil {
			// done replaying log; new data incoming
			// OR signaled to load snapshot
			snapshot, err := s.snapshots.Load()
			if err == snap.ErrNoSnapshot {
				return
			}
			if err != nil {
				log.Panic(err)
			}
			log.Printf("loading snapshot at term %d and index %d", snapshot.Metadata.Term, snapshot.Metadata.Index)
			if err := s.recoverFromSnapshot(snapshot.Data); err != nil {
				log.Panic(err)
			}
			continue
		}

		var dataKv kv
		dec := gob.NewDecoder(bytes.NewBufferString(*data))
		if err := dec.Decode(&dataKv); err != nil {
			log.Fatalf("raftexample: could not decode message (%v)", err)
		}
		s.mu.Lock()
		s.kvStore[dataKv.Key] = dataKv.Val
		s.mu.Unlock()
	}
	if err, ok := <-errorC; ok {
		log.Fatal(err)
	}
}

func (s *keystore) getSnapshot() ([]byte, error) {
	return make([]byte, 0), nil
	//s.mu.RLock()
	//defer s.mu.RUnlock()
	//return json.Marshal(s.kvStore)
}

func (s *keystore) recoverFromSnapshot(snapshot []byte) error {
	//var store map[string]string
	//if err := json.Unmarshal(snapshot, &store); err != nil {
	//	return err
	//}
	//s.mu.Lock()
	//defer s.mu.Unlock()
	//s.kvStore = store
	return nil
}
