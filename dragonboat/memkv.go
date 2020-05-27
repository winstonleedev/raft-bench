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
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
	"unsafe"

	sm "github.com/lni/dragonboat/v3/statemachine"
)

const (
	appliedIndexKey    string = "disk_kv_applied_index"
	testDBDirName      string = "example-data"
	currentDBFilename  string = "current"
	updatingDBFilename string = "current.updating"
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
	Key string
	Val string
}

// functions below are used to manage the current data directory of RocksDB DB.
func isNewRun(dir string) bool {
	fp := filepath.Join(dir, currentDBFilename)
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		return true
	}
	return false
}

func getNodeDBDirName(clusterID uint64, nodeID uint64) string {
	part := fmt.Sprintf("%d_%d", clusterID, nodeID)
	return filepath.Join(testDBDirName, part)
}

func getNewRandomDBDirName(dir string) string {
	part := "%d_%d"
	rn := rand.Uint64()
	ct := time.Now().UnixNano()
	return filepath.Join(dir, fmt.Sprintf(part, rn, ct))
}

func replaceCurrentDBFile(dir string) error {
	fp := filepath.Join(dir, currentDBFilename)
	tmpFp := filepath.Join(dir, updatingDBFilename)
	if err := os.Rename(tmpFp, fp); err != nil {
		return err
	}
	return syncDir(dir)
}

func saveCurrentDBDirName(dir string, dbdir string) error {
	h := md5.New()
	if _, err := h.Write([]byte(dbdir)); err != nil {
		return err
	}
	fp := filepath.Join(dir, updatingDBFilename)
	f, err := os.Create(fp)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
		if err := syncDir(dir); err != nil {
			panic(err)
		}
	}()
	if _, err := f.Write(h.Sum(nil)[:8]); err != nil {
		return err
	}
	if _, err := f.Write([]byte(dbdir)); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	return nil
}

func getCurrentDBDirName(dir string) (string, error) {
	fp := filepath.Join(dir, currentDBFilename)
	f, err := os.OpenFile(fp, os.O_RDONLY, 0755)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	if len(data) <= 8 {
		panic("corrupted content")
	}
	crc := data[:8]
	content := data[8:]
	h := md5.New()
	if _, err := h.Write(content); err != nil {
		return "", err
	}
	if !bytes.Equal(crc, h.Sum(nil)[:8]) {
		panic("corrupted content with not matched crc")
	}
	return string(content), nil
}

func createNodeDataDir(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return syncDir(filepath.Dir(dir))
}

func cleanupNodeDataDir(dir string) error {
	os.RemoveAll(filepath.Join(dir, updatingDBFilename))
	dbdir, err := getCurrentDBDirName(dir)
	if err != nil {
		return err
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, fi := range files {
		if !fi.IsDir() {
			continue
		}
		fmt.Printf("dbdir %s, fi.name %s, dir %s\n", dbdir, fi.Name(), dir)
		toDelete := filepath.Join(dir, fi.Name())
		if toDelete != dbdir {
			fmt.Printf("removing %s\n", toDelete)
			if err := os.RemoveAll(toDelete); err != nil {
				return err
			}
		}
	}
	return nil
}

// MemKV is a state machine that implements the IOnDiskStateMachine interface.
// MemKV stores key-value pairs in the underlying RocksDB key-value store. As
// it is used as an example, it is implemented using the most basic features
// common in most key-value stores. This is NOT a benchmark program.
type MemKV struct {
	clusterID   uint64
	nodeID      uint64
	lastApplied uint64
	db          unsafe.Pointer
	closed      bool
	aborted     bool

	mu      sync.RWMutex
	kvStore map[interface{}]string // current committed key-value pairs
}

// NewMemKV creates a new disk kv test state machine.
func NewMemKV(clusterID uint64, nodeID uint64) sm.IOnDiskStateMachine {
	d := &MemKV{
		clusterID: clusterID,
		nodeID:    nodeID,
		kvStore:   make(map[interface{}]string),
	}
	return d
}

func (d *MemKV) queryAppliedIndex() (uint64, error) {
	return 0, nil
}

// Open opens the state machine and return the index of the last Raft Log entry
// already updated into the state machine.
func (d *MemKV) Open(stopc <-chan struct{}) (uint64, error) {
	dir := getNodeDBDirName(d.clusterID, d.nodeID)
	if err := createNodeDataDir(dir); err != nil {
		panic(err)
	}
	var dbdir string
	if !isNewRun(dir) {
		if err := cleanupNodeDataDir(dir); err != nil {
			return 0, err
		}
		var err error
		dbdir, err = getCurrentDBDirName(dir)
		if err != nil {
			return 0, err
		}
		if _, err := os.Stat(dbdir); err != nil {
			if os.IsNotExist(err) {
				panic("db dir unexpectedly deleted")
			}
		}
	} else {
		dbdir = getNewRandomDBDirName(dir)
		if err := saveCurrentDBDirName(dir, dbdir); err != nil {
			return 0, err
		}
		if err := replaceCurrentDBFile(dir); err != nil {
			return 0, err
		}
	}
	return 0, nil
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
func (d *MemKV) Update(ents []sm.Entry) ([]sm.Entry, error) {
	if d.aborted {
		panic("update() called after abort set to true")
	}
	if d.closed {
		panic("update called after Close()")
	}
	for idx, e := range ents {
		dataKV := &KVData{}
		if err := json.Unmarshal(e.Cmd, dataKV); err != nil {
			panic(err)
		}
		d.kvStore[dataKV.Key] = dataKV.Val
		ents[idx].Result = sm.Result{Value: uint64(len(ents[idx].Cmd))}
	}
	// save the applied index to the DB.
	if d.lastApplied >= ents[len(ents)-1].Index {
		panic("lastApplied not moving forward")
	}
	d.lastApplied = ents[len(ents)-1].Index
	return ents, nil
}

// Sync synchronizes all in-core state of the state machine. Since the Update
// method in this example already does that every time when it is invoked, the
// Sync method here is a NoOP.
func (d *MemKV) Sync() error {
	return nil
}

// PrepareSnapshot prepares snapshotting. PrepareSnapshot is responsible to
// capture a state identifier that identifies a point in time state of the
// underlying data. In this example, we use RocksDB's snapshot feature to
// achieve that.
func (d *MemKV) PrepareSnapshot() (interface{}, error) {
	if d.closed {
		panic("prepare snapshot called after Close()")
	}
	if d.aborted {
		panic("prepare snapshot called after abort")
	}
	return d, nil
}

// saveToWriter saves all existing key-value pairs to the provided writer.
// As an example, we use the most straight forward way to implement this.
func (d *MemKV) saveToWriter(w io.Writer) error {
	sz := make([]byte, 8)
	binary.LittleEndian.PutUint64(sz, uint64(len(d.kvStore)))
	if _, err := w.Write(sz); err != nil {
		return err
	}
	for key, val := range d.kvStore {
		dataKv := &KVData{
			Key: fmt.Sprintf("%v", key),
			Val: val,
		}
		data, err := json.Marshal(dataKv)
		if err != nil {
			panic(err)
		}
		binary.LittleEndian.PutUint64(sz, uint64(len(data)))
		if _, err := w.Write(sz); err != nil {
			return err
		}
		if _, err := w.Write(data); err != nil {
			return err
		}
	}
	return nil
}

// SaveSnapshot saves the state machine state identified by the state
// identifier provided by the input ctx parameter. Note that SaveSnapshot
// is not suppose to save the latest state.
func (d *MemKV) SaveSnapshot(ctx interface{},
	w io.Writer, done <-chan struct{}) error {
	if d.closed {
		panic("prepare snapshot called after Close()")
	}
	if d.aborted {
		panic("prepare snapshot called after abort")
	}
	return d.saveToWriter(w)
}

// RecoverFromSnapshot recovers the state machine state from snapshot. The
// snapshot is recovered into a new DB first and then atomically swapped with
// the existing DB to complete the recovery.
func (d *MemKV) RecoverFromSnapshot(r io.Reader,
	done <-chan struct{}) error {
	if d.closed {
		panic("recover from snapshot called after Close()")
	}
	dir := getNodeDBDirName(d.clusterID, d.nodeID)
	dbdir := getNewRandomDBDirName(dir)
	oldDirName, err := getCurrentDBDirName(dir)
	if err != nil {
		return err
	}
	sz := make([]byte, 8)
	if _, err := io.ReadFull(r, sz); err != nil {
		return err
	}
	total := binary.LittleEndian.Uint64(sz)
	for i := uint64(0); i < total; i++ {
		if _, err := io.ReadFull(r, sz); err != nil {
			return err
		}
		toRead := binary.LittleEndian.Uint64(sz)
		data := make([]byte, toRead)
		if _, err := io.ReadFull(r, data); err != nil {
			return err
		}
		dataKv := &KVData{}
		if err := json.Unmarshal(data, dataKv); err != nil {
			panic(err)
		}
		d.kvStore[dataKv.Key] = dataKv.Val
	}
	if err := saveCurrentDBDirName(dir, dbdir); err != nil {
		return err
	}
	if err := replaceCurrentDBFile(dir); err != nil {
		return err
	}
	newLastApplied, err := d.queryAppliedIndex()
	if err != nil {
		panic(err)
	}
	if d.lastApplied > newLastApplied {
		panic("last applied not moving forward")
	}
	d.lastApplied = newLastApplied
	return os.RemoveAll(oldDirName)
}

// Close closes the state machine.
func (d *MemKV) Close() error {
	return nil
}

// GetHash returns a hash value representing the state of the state machine.
func (d *MemKV) GetHash() (uint64, error) {
	h := md5.New()
	md5sum := h.Sum(nil)
	return binary.LittleEndian.Uint64(md5sum[:8]), nil
}
