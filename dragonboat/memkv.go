// Copyright 2017-2019 Lei Ni (nilei81@gmail.com)
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

package dragonboat

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"

	sm "github.com/lni/dragonboat/v3/statemachine"
)

func syncDir(dir string) (err error) {
	if runtime.GOOS == "windows" {
		return nil
	}
	fileInfo, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !fileInfo.IsDir() {
		panic("not a dir")
	}
	df, err := os.Open(filepath.Clean(dir))
	if err != nil {
		return err
	}
	defer func() {
		if cerr := df.Close(); err == nil {
			err = cerr
		}
	}()
	return df.Sync()
}

type KVData struct {
	Key interface{}
	Val interface{}
}

// MemKV is a state machine that implements the IOnDiskStateMachine interface.
// MemKV stores key-value pairs in the underlying RocksDB key-value store. As
// it is used as an example, it is implemented using the most basic features
// common in most key-value stores. This is NOT a benchmark program.
type MemKV struct {
	clusterID uint64
	nodeID    uint64
	db        unsafe.Pointer
	kvStore   map[interface{}]interface{} // current committed key-value pairs
}

// NewMemKV creates a new disk kv test state machine.
func NewMemKV(clusterID uint64, nodeID uint64) sm.IStateMachine {
	d := &MemKV{
		clusterID: clusterID,
		nodeID:    nodeID,
		kvStore:   make(map[interface{}]interface{}),
	}
	return d
}

// Lookup queries the state machine.
func (d *MemKV) Lookup(key interface{}) (interface{}, error) {
	return d.kvStore[key], nil
}

// Update updates the state machine. In this example, all updates are put into
// a RocksDB write batch and then atomically written to the DB together with
// the index of the last Raft Log entry. For simplicity, we always Sync the
// writes (db.wo.Sync=True). To get higher throughput, you can implement the
// Sync() method below and choose not to synchronize for every Update(). Sync()
// will periodically called by Dragonboat to synchronize the state.
func (d *MemKV) Update(data []byte) (sm.Result, error) {
	kv := &KVData{}
	err := json.Unmarshal(data, kv)
	if err == nil {
		d.kvStore[kv.Key] = kv.Val
	}
	return sm.Result{Value: uint64(len(data)), Data: data}, err
}

// SaveSnapshot saves the state machine state identified by the state
// identifier provided by the input ctx parameter. Note that SaveSnapshot
// is not suppose to save the latest state.
func (d *MemKV) SaveSnapshot(w io.Writer,
	fileCollection sm.ISnapshotFileCollection,
	done <-chan struct{}) error {
	data, err := json.Marshal(d)
	if err != nil {
		panic(err)
	}
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// RecoverFromSnapshot recovers the state machine state from snapshot. The
// snapshot is recovered into a new DB first and then atomically swapped with
// the existing DB to complete the recovery.
func (d *MemKV) RecoverFromSnapshot(r io.Reader,
	files []sm.SnapshotFile,
	done <-chan struct{}) error {
	var sn MemKV
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &sn)
	if err != nil {
		panic("failed to unmarshal snapshot")
	}

	return nil
}

// Close closes the state machine.
func (d *MemKV) Close() error {
	return nil
}

// GetHash returns a hash value representing the state of the state machine.
func (d *MemKV) GetHash() (uint64, error) {
	h := md5.New()
	data, _ := json.Marshal(d)
	md5sum := h.Sum(data)
	return binary.LittleEndian.Uint64(md5sum[:8]), nil
}
