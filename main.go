package main

import (
	"flag"

	"github.com/thanhphu/raftbench/dragonboat"
	"github.com/thanhphu/raftbench/etcd"
	"github.com/thanhphu/raftbench/hashicorp"
)

func main() {
	engine := flag.String("engine", "etcd", "etcd/hashicorp/dragonboat select the raft engine")
	switch *engine {
	case "etcd":
		etcd.Main()
	case "hashicorp":
		hashicorp.Main()
	case "dragonboat":
		dragonboat.Main()
	}
}
