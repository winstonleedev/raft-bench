package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/thanhphu/raftbench/dragonboat"
	"github.com/thanhphu/raftbench/etcd"
	"github.com/thanhphu/raftbench/hashicorp"
	"github.com/thanhphu/raftbench/util"
)

// Command line defaults
const (
	DefaultHTTPAddr = ":11000"
	DefaultRaftAddr = ":12000"
)

func main() {
	engine := flag.String("engine", "etcd", "etcd/hashi/dragonboat select the raft engine")
	cluster := flag.String("cluster", "http://127.0.0.1:9021", "comma separated cluster peers")
	id := flag.Int("id", 1, "node ID")
	kvPort := flag.Int("port", 9121, "key-value server port")
	join := flag.Bool("join", false, "join an existing cluster")
	enabled := flag.Bool("test", false, "use this node for r/w tests")
	logFile := flag.String("logfile", "result.csv", "name of csv log file (used together with --test)")
	numKeys := flag.Int("numKeys", 1, "Run the benchmark numKeys * mil times")
	mil := flag.Int("mil", 1000000, "Run the benchmark numKeys * mil times")
	runs := flag.Int("runs", 10, "Number of time to run the benchmark")
	wait := flag.Int("wait", 3000, "Time to wait before each step and before read / write")
	firstWait := flag.Int("firstWait", 10000, "Time to wait before starting benchmark")
	step := flag.Int("step", 100, "If read fails, wait this much before trying to avoid overloading the system")
	maxTries := flag.Int("maxTries", 10, "Only retry an operation this many times")

	var httpAddr string
	var raftAddr string
	var joinAddr string
	var nodeID string

	flag.StringVar(&httpAddr, "haddr", DefaultHTTPAddr, "Set the HTTP bind address")
	flag.StringVar(&raftAddr, "raddr", DefaultRaftAddr, "Set Raft bind address")
	flag.StringVar(&joinAddr, "joinaddr", "", "Set join address, if any")
	flag.StringVar(&nodeID, "nodeid", "", "Node ID")

	addr := flag.String("addr", "", "Nodehost address")

	flag.Usage = func() {
		log.Printf("Usage: %s [options] <raft-data-path> \n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	testParams := &util.TestParams{
		NumKeys:   *numKeys,
		Mil:       *mil,
		Runs:      *runs,
		Wait:      time.Duration(*wait) * time.Millisecond,
		FirstWait: time.Duration(*firstWait) * time.Millisecond,
		Step:      time.Duration(*step) * time.Millisecond,
		MaxTries:  *maxTries,
		Enabled:   *enabled,
		LogFile:   *logFile,
	}

	switch *engine {
	case "etcd":
		etcd.Main(*cluster, *id, *kvPort, *join, *testParams)
	case "hashi":
		hashicorp.Main(httpAddr, raftAddr, joinAddr, nodeID, *testParams)
	case "dragonboat":
		dragonboat.Main(*id, *addr, *join, *testParams)
	}
}
