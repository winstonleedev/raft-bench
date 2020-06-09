package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/thanhphu/raftbench/dragonboat"
	"github.com/thanhphu/raftbench/etcd"
	"github.com/thanhphu/raftbench/hashicorp"
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
	kvport := flag.Int("port", 9121, "key-value server port")
	join := flag.Bool("join", false, "join an existing cluster")
	test := flag.Bool("test", false, "use this node for r/w tests")

	var inmem bool
	var httpAddr string
	var raftAddr string
	var joinAddr string
	var nodeID string

	flag.BoolVar(&inmem, "inmem", false, "Use in-memory storage for Raft")
	flag.StringVar(&httpAddr, "haddr", DefaultHTTPAddr, "Set the HTTP bind address")
	flag.StringVar(&raftAddr, "raddr", DefaultRaftAddr, "Set Raft bind address")
	flag.StringVar(&joinAddr, "joinport", "", "Set join address, if any")
	flag.StringVar(&nodeID, "nodeid", "", "Node ID")

	addr := flag.String("addr", "", "Nodehost address")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <raft-data-path> \n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	switch *engine {
	case "etcd":
		etcd.Main(*cluster, *id, *kvport, *join, *test)
	case "hashi":
		hashicorp.Main(inmem, httpAddr, raftAddr, joinAddr, nodeID, *test)
	case "dragonboat":
		dragonboat.Main(*id, *addr, *join, *test)
	}
}
