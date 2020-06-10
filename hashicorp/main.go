package hashicorp

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/thanhphu/raftbench/hashicorp/http"
	"github.com/thanhphu/raftbench/hashicorp/store"
	"github.com/thanhphu/raftbench/util"
)

func Main(httpAddr string, raftAddr string, joinAddr string, nodeID string, test bool, logFile string) {
	if flag.NArg() == 0 {
		log.Printf("No Raft storage directory specified\n")
		os.Exit(1)
	}

	// Ensure Raft storage exists.
	raftDir := flag.Arg(0)
	if raftDir == "" {
		log.Printf("No Raft storage directory specified\n")
		os.Exit(1)
	}
	err := os.MkdirAll(raftDir, 0700)
	if err != nil {
		log.Printf("Unable to create WAL directory\n")
		os.Exit(1)
	}

	s := store.New(true)
	s.RaftDir = raftDir
	s.RaftBind = raftAddr
	if err := s.Open(joinAddr == "", nodeID); err != nil {
		log.Fatalf("failed to open store: %s", err.Error())
	}

	h := httpd.New(httpAddr, s)
	if err := h.Start(); err != nil {
		log.Fatalf("failed to start HTTP service: %s", err.Error())
	}

	// If join was specified, make the join request.
	if joinAddr != "" {
		if err := join(joinAddr, raftAddr, nodeID); err != nil {
			log.Fatalf("failed to join node at %s: %s", joinAddr, err.Error())
		}
	}

	log.Println("raftbench started successfully")

	util.Bench(test, logFile, func(k string) bool {
		_, err := s.Get(k)
		if err != nil {
			// log.Printf("error retrieving key %v\n", err)
			return false
		}
		return true
	}, func(k string, v string) bool {
		err := s.Set(k, v)
		if err != nil {
			// log.Printf("error setting key %v\n", err)
			return false
		}
		return true
	})
}

func join(joinAddr, raftAddr, nodeID string) error {
	b, err := json.Marshal(map[string]string{"addr": raftAddr, "id": nodeID})
	if err != nil {
		return err
	}
	resp, err := http.Post(fmt.Sprintf("http://%s/join", joinAddr), "application-type/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer func() {
		if resp.Body.Close() != nil {
			log.Printf("Error closing")
		}
	}()

	return nil
}
