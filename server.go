// server.go
package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	// "sync"
	"time"
)

// ServerRPC defines server-side RPC methods
// type ServerRPC struct {
// 	Nodes map[int]*HeartbeatTable
// 	Mutex sync.Mutex
// }


// CleanupFailedNodes removes nodes that have been marked as failed for too long
func (s *ServerRPC) CleanupNodes() {
    for id, entry := range s.Membership {
        if entry.Failed {
            fmt.Printf("Server is removing Node %d from the global membership list\n", id)
            delete(s.Membership, id)
        }
    }
}


// RegisterNode registers a new node with the server
func (s *ServerRPC) RegisterNode(args *RegisterNodeArgs, reply *RegisterNodeReply) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if _, exists := s.Nodes[args.ID]; !exists {
		s.Nodes[args.ID] = &HeartbeatTable{Entries: make(map[int]HeartbeatEntry)}
		s.PeerAddresses[args.ID] = fmt.Sprintf("localhost:%d", 1234+args.ID) // assign unique port for each client
		
		s.Membership[args.ID] = HeartbeatEntry {
			ID: args.ID,
			Counter: 0,
			Timestamp: time.Now(),
		}
		
		reply.Success = true
		fmt.Printf("Node %d registered successfully\n", args.ID)
	} else {
		reply.Success = false
		fmt.Printf("Node %d already registered\n", args.ID)
	}

	// update peer list for all existing nodes
	for id, addr := range s.PeerAddresses {
		if id != args.ID {
			reply.Peers = append(reply.Peers, addr)
		}
	}

	// Provide the full membership list to the new node
	reply.Peers = make([]string, 0)
	for id, addr := range s.PeerAddresses {
		if id != args.ID {
			reply.Peers = append(reply.Peers, addr)
		}
	}
	return nil
}

// GetMembershipList returns the current global membership list
func (s *ServerRPC) GetMembershipList(args *int, reply *map[int]HeartbeatEntry) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	*reply = make(map[int]HeartbeatEntry)
	for id, entry := range s.Membership {
		if !entry.Failed {
			(*reply)[id] = entry
		}	
	}

	return nil
}


// Gossip updates the server with a client's heartbeat table
func (s *ServerRPC) Gossip(msg *GossipMessage, reply *bool) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if nodeTable, exists := s.Nodes[msg.SenderID]; exists {
		nodeTable.Mutex.Lock()
		for id, entry := range msg.Table {
			// Only update if the received entry is more recent

			if existing, exists := s.Membership[id]; exists && existing.Failed {
				continue
			}

			if existing, exists := s.Membership[id]; !exists || entry.Counter > existing.Counter {

				if time.Since(entry.Timestamp) > FAILURE_TIMEOUT {
                    fmt.Printf("Server detected failure of Node %d (Last update: %s)\n", id, entry.Timestamp.Format(time.RFC3339))
                    entry.Failed = true
                }

				s.Membership[id] = entry
			}
			nodeTable.Entries[id] = entry
		}
		nodeTable.Mutex.Unlock()
		s.CleanupNodes()
		nodeTable.Print()
		*reply = true
		fmt.Printf("Received gossip from Node %d\n", msg.SenderID)
	} else {
		*reply = false
		fmt.Printf("Gossip received from unknown node %d\n", msg.SenderID)
	}
	return nil
}

func startServer() {
	server := &ServerRPC {
		Nodes: make(map[int]*HeartbeatTable),
		PeerAddresses: make(map[int]string),
		Membership: make(map[int]HeartbeatEntry),
	}
	
	rpc.Register(server)

	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("Listen error:", err)
	}
	fmt.Println("Server listening on port 1234")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Connection error:", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}

func main() {
	startServer()

}