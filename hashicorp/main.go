package hashicorp

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/thanhphu/raftbench/hashicorp/http"
	"github.com/thanhphu/raftbench/hashicorp/store"
)

const (
	numKeys = 1
	mil     = 1000000
)

func Main(inmem bool, httpAddr string, raftAddr string, joinAddr string, nodeID string, test bool) {
	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "No Raft storage directory specified\n")
		os.Exit(1)
	}

	// Ensure Raft storage exists.
	raftDir := flag.Arg(0)
	if raftDir == "" {
		fmt.Fprintf(os.Stderr, "No Raft storage directory specified\n")
		os.Exit(1)
	}
	os.MkdirAll(raftDir, 0700)

	s := store.New(inmem)
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

	log.Println("hraftd started successfully")

	if test {
		for i := 0; i < 3; i++ {
			time.Sleep(3000)

			start := time.Now()
			k := 0
			for k < numKeys*mil {
				v := rand.Int()
				go s.Set(string(k), string(v))
				k += 1
			}
			fmt.Printf("Write test, %v, %v, %v\n", i+1, numKeys*mil, time.Since(start))

			time.Sleep(3000)
			start = time.Now()
			k = 0
			for k < numKeys*mil {
				go s.Get(string(k))
				k += 1
			}
			fmt.Printf("Read test, %v, %v, %v\n", i+1, numKeys*mil, time.Since(start))
		}
	}

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	log.Println("hraftd exiting")
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
	defer resp.Body.Close()

	return nil
}
