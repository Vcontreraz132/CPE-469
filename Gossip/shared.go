// shared.go
package main

import (
	"sync"
	"time"
	"fmt"
	"net/rpc"
)

// HeartbeatEntry represents an entry in the heartbeat table
type HeartbeatEntry struct {
	ID        int
	Neighbor  int
	Counter   int
	Timestamp time.Time
}

// HeartbeatTable stores heartbeat information of a node
type HeartbeatTable struct {
	Entries map[int]HeartbeatEntry
	Mutex   sync.Mutex
}

// GossipMessage represents the data sent during gossip
type GossipMessage struct {
	SenderID  int
	Table     map[int]HeartbeatEntry
}

// ServerRPC defines server-side RPC methods
type ServerRPC struct {
	Nodes map[int]*HeartbeatTable
	PeerAddresses map[int]string // store registered peer addresses
	Membership map[int]HeartbeatEntry // Global membership list
	Mutex sync.Mutex
}

// ClientRPC defines client-side RPC methods
type ClientRPC struct {
	ID     int
	Table  HeartbeatTable
	Mutex  sync.Mutex
	Peers []*rpc.Client // list of peers for gossip
}

// RegisterNodeArgs is used when clients register with the server
type RegisterNodeArgs struct {
	ID int
}

// RegisterNodeReply is the response from the server
type RegisterNodeReply struct {
	Success bool
	Peers []string // list of peer addresses
}

func (h* HeartbeatTable) Print() {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()
	fmt.Println("Heartbeat Table:")
	for id, entry := range h.Entries {
		fmt.Printf("Node %d -> Counter: %d, Timestamp: %s\n", id, entry.Counter, entry.Timestamp.Format(time.RFC3339))
	}
}
