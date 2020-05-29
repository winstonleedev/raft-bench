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
	engine := flag.String("engine", "etcd", "etcd/hashicorp/dragonboat select the raft engine")
	cluster := flag.String("cluster", "http://127.0.0.1:9021", "comma separated cluster peers")
	id := flag.Int("id", 1, "node ID")
	kvport := flag.Int("port", 9121, "key-value server port")
	join := flag.Bool("join", false, "join an existing cluster")

	var inmem bool
	var httpAddr string
	var raftAddr string
	var joinAddr string
	var nodeID string

	flag.BoolVar(&inmem, "inmem", false, "Use in-memory storage for Raft")
	flag.StringVar(&httpAddr, "haddr", DefaultHTTPAddr, "Set the HTTP bind address")
	flag.StringVar(&raftAddr, "raddr", DefaultRaftAddr, "Set Raft bind address")
	flag.StringVar(&joinAddr, "join2", "", "Set join address, if any")
	flag.StringVar(&nodeID, "id2", "", "Node ID")

	nodeID2 := flag.Int("nodeid2", 1, "NodeID to use")
	addr2 := flag.String("addr2", "", "Nodehost address")
	join2 := flag.Bool("join3", false, "Joining a new node")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <raft-data-path> \n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	switch *engine {
	case "etcd":
		etcd.Main(*cluster, *id, *kvport, *join)
	case "hashicorp":
		hashicorp.Main(inmem, httpAddr, raftAddr, joinAddr, nodeID)
	case "dragonboat":
		dragonboat.Main(*nodeID2, *addr2, *join2)
	}
}

func join(joinAddr, raftAddr, nodeID string) error {

	return nil
}
