package main

import (
	"fmt"
	"log"
	"math/rand"
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

// Waits for the specified delay and then removes the node with the given ID.
func (s *ServerRPC) FailOneNodeAfter(nodeID int, delay time.Duration) {
	time.Sleep(delay)

	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if _, exists := s.Membership[nodeID]; exists {
		fmt.Printf("Server: Node %d has failed after %v. Removing from membership.\n", nodeID, delay)
		delete(s.Membership, nodeID)
		delete(s.Nodes, nodeID)
		delete(s.PeerAddresses, nodeID)
	} else {
		fmt.Printf("Server: Node %d not found in membership. Nothing to remove.\n", nodeID)
	}
}

// Randomly fails one node every `interval` seconds.
func (s *ServerRPC) SimulateFailures(interval time.Duration) {
	for {
		time.Sleep(interval)

		s.Mutex.Lock()
		// If there are no nodes registered, skip.
		if len(s.Membership) == 0 {
			s.Mutex.Unlock()
			continue
		}

		// Build a slice of node IDs from the membership list.
		var nodeIDs []int
		for id := range s.Membership {
			nodeIDs = append(nodeIDs, id)
		}

		// Randomly pick one node ID to "fail"
		failedNode := nodeIDs[rand.Intn(len(nodeIDs))]

		// Remove the failed node from the membership list and other maps.
		fmt.Printf("Server: Node %d has failed! Removing from membership list.\n", failedNode)
		delete(s.Membership, failedNode)
		delete(s.Nodes, failedNode)
		delete(s.PeerAddresses, failedNode)
		s.Mutex.Unlock()
	}
}

// RegisterNode registers a new node with the server
func (s *ServerRPC) RegisterNode(args *RegisterNodeArgs, reply *RegisterNodeReply) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if _, exists := s.Nodes[args.ID]; !exists {
		s.Nodes[args.ID] = &HeartbeatTable{Entries: make(map[int]HeartbeatEntry)}
		s.PeerAddresses[args.ID] = fmt.Sprintf("localhost:%d", 1234+args.ID) // assign unique port for each client

		s.Membership[args.ID] = HeartbeatEntry{
			ID:        args.ID,
			Counter:   0,
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
		(*reply)[id] = entry
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
			if existing, exists := s.Membership[id]; !exists || entry.Counter > existing.Counter {
				s.Membership[id] = entry
			}
			nodeTable.Entries[id] = entry
		}
		nodeTable.Mutex.Unlock()
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
	// Initialize & declare server struct //
	server := &ServerRPC{
		Nodes:         make(map[int]*HeartbeatTable),
		PeerAddresses: make(map[int]string),
		Membership:    make(map[int]HeartbeatEntry),
	}

	rpc.Register(server)

	// Listen for connections //
	listener, err := net.Listen("tcp", ":1234") // Create an object for listening
	if err != nil {
		log.Fatal("Listen error:", err)
	}
	fmt.Println("Server listening on port 1234")

	// // Start the failure simulation every 10 seconds.
	// go server.SimulateFailures(10 * time.Second)

	// Launch the one-time failure simulation.
	go server.FailOneNodeAfter(3, 15*time.Second)

	// Accept incoming connections //
	for {
		conn, err := listener.Accept() // Create an object for incoming connections
		if err != nil {
			log.Println("Connection error:", err)
			continue
		}
		go rpc.ServeConn(conn) // Accept connection
	}
}

func main() {
	startServer()
}
