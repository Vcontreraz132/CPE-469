// client.go
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"net/rpc"
	"sync"
	"time"
	"net"
	"sort"
)

// Client represents a node in the gossip network
type Client struct {
	ID     int
	Table  HeartbeatTable
	Mutex  sync.Mutex
	Server *rpc.Client
	Peers []*rpc.Client
	PeerAddresses map[int]string // Store connected peer addresses
}

// ListenForGossip allows the client to receive gossip messages
func (c *Client) ListenForGossip(port int) {
	rpc.Register(c)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal("Listen error:", err)
	}
	fmt.Printf("Node %d listening on port %d for peer gossip\n", c.ID, port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Connection error:", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}

// Gossip handles incoming gossip messages
func (c *Client) Gossip(msg *GossipMessage, reply *bool) error {
	c.Mutex.Lock()
	// Merge received gossip data into own table
	for id, entry := range msg.Table {
		// If entry is newer, update our table
		if existing, exists := c.Table.Entries[id]; exists {
			if existing.Failed {
                continue // Ignore failed nodes
            }
			if entry.Counter > existing.Counter {
                c.Table.Entries[id] = entry
            }
		} else {
			if !entry.Failed { // Prevent reintroducing failed nodes
                c.Table.Entries[id] = entry
            }
		}
	}
	c.Mutex.Unlock()
	fmt.Printf("Node %d received gossip from Node %d\n", c.ID, msg.SenderID)
	*reply = true
	return nil
}

// StartHeartbeat increments the heartbeat counter every X seconds
func (c *Client) StartHeartbeat(interval time.Duration) {
	for {
		time.Sleep(interval)
		c.Mutex.Lock()
		c.Table.Entries[c.ID] = HeartbeatEntry{
			ID:        c.ID,
			Neighbor:  c.ID, // Self-referential initially
			Counter:   c.Table.Entries[c.ID].Counter + 1,
			Timestamp: time.Now(),
		}
		c.Mutex.Unlock()
		fmt.Printf("Node %d heartbeat incremented", c.ID)
		c.PrintTable()
	}
}

// SendGossip sends heartbeat data to the server and peers
func (c *Client) SendGossip(interval time.Duration) {
	for {
		time.Sleep(interval)
		c.Mutex.Lock()
		msg := GossipMessage{SenderID: c.ID, Table: c.Table.Entries}
		for id, entry := range c.Table.Entries {
			if !entry.Failed {
				msg.Table[id] = entry
			}
		}
		c.Mutex.Unlock()

		var reply bool
		// Send gossip to the server
		if err := c.Server.Call("ServerRPC.Gossip", &msg, &reply); err != nil {
			log.Println("Error sending gossip:", err)
		} else {
			fmt.Printf("Node %d sent gossip to the server\n", c.ID)
		}

		// update membership list
		c.UpdateMembershipList()

		// Select peers from the updated membership list
		for id, addr := range c.PeerAddresses {
			if id != c.ID { // Avoid self-gossip
				peer, err := rpc.Dial("tcp", addr)
				if err == nil {
					defer peer.Close()
					if err := peer.Call("Client.Gossip", &msg, &reply); err != nil {
						fmt.Printf("Error sending gossip to Node %d: %v\n", id, err)
					} else {
						fmt.Printf("Node %d sent gossip to Node %d\n", c.ID, id)
					}
				}
			}
		}

		c.PrintTable()
	}
}


// PrintTable prints the client's current heartbeat table in ascending order of node ID
func (c *Client) PrintTable() {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	// Extract node IDs and sort them
	ids := make([]int, 0, len(c.Table.Entries))
	for id := range c.Table.Entries {
		ids = append(ids, id)
	}
	sort.Ints(ids) // Sort IDs in ascending order

	// Print the table header
	fmt.Printf("\nNode %d Heartbeat Table:\n", c.ID)

	// Print each node's heartbeat entry in sorted order
	for _, id := range ids {
		entry := c.Table.Entries[id]
		fmt.Printf("Node %d -> Counter: %d, Timestamp: %s\n", id, entry.Counter, entry.Timestamp.Format(time.RFC3339))
	}
}

// UpdatePeerList allows the server to notify clients of new peers
func (c *Client) UpdatePeerList(peers map[int]string, reply *bool) error {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	// ensure PeerAddresses is initialized before use
	if c.PeerAddresses == nil {
		c.PeerAddresses = make(map[int]string)
	}

	for id, addr := range peers {
		if id != c.ID {
			// Check if the peer is already connected
			if _, exists := c.PeerAddresses[id]; !exists {
				peer, err := rpc.Dial("tcp", addr)
				if err == nil {
					c.Peers = append(c.Peers, peer)
					c.PeerAddresses[id] = addr // Track connected peer
					fmt.Printf("Node %d discovered and connected to new peer at %s\n", c.ID, addr)
				}
			}
		}
	}
	*reply = true
	return nil
}

// Request and update the global membership list from the server
func (c *Client) UpdateMembershipList() {
	var membershipList map[int]HeartbeatEntry
	err := c.Server.Call("ServerRPC.GetMembershipList", &c.ID, &membershipList)
	if err != nil {
		fmt.Println("Error requesting membership list:", err)
		return
	}

	// Update local table with global membership data
	c.Mutex.Lock()
	for id, entry := range membershipList {
		c.Table.Entries[id] = entry
	}
	c.Mutex.Unlock()
	fmt.Println("Updated global membership list")
}

// CheckNodeFailures periodically scans the heartbeat table and marks nodes as failed
func (c *Client) CheckNodeFailures(interval time.Duration) {
	for {
		time.Sleep(interval)

		c.Mutex.Lock()
		currentTime := time.Now()

		for id, entry := range c.Table.Entries {
			if id == c.ID {
				continue // Skip self
			}

			// Check if the last timestamp is too old
			if !entry.Failed && currentTime.Sub(entry.Timestamp) > FAILURE_TIMEOUT {
                fmt.Printf("Warning: Node %d detected failure of Node %d (Last update: %s)\n", c.ID, id, entry.Timestamp.Format(time.RFC3339))
                c.Table.Entries[id] = HeartbeatEntry{
                    ID:        id,
                    Counter:   entry.Counter,
                    Timestamp: entry.Timestamp,
                    Failed:    true, // Mark as failed
                }
            }
        }
        c.Mutex.Unlock()
    }
}


func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run client.go <node_id>")
	}
	clientID, err := strconv.Atoi(os.Args[1])
	if err != nil || clientID < 1 || clientID > 8 {
		log.Fatal("Invalid Node ID. Must be an integer between 1 and 8.")
	}

	server, err := rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("Connection error:", err)
	}

	client := &Client{
		ID:     clientID,
		Table:  HeartbeatTable{Entries: make(map[int]HeartbeatEntry)},
		Server: server,
		Peers: []*rpc.Client{},
		PeerAddresses: make(map[int]string),
	}

	// Register client with server
	args := RegisterNodeArgs{ID: clientID}
	var reply RegisterNodeReply
	if err := server.Call("ServerRPC.RegisterNode", &args, &reply); err != nil || !reply.Success {
		log.Fatal("Registration failed")
	}
	fmt.Printf("Node %d registered successfully", clientID)

	// Start listening for peer gossip on a unique port
	peerPort := 1234 + clientID
	go client.ListenForGossip(peerPort)


	// Request the full membership list from the server
	var membershipList map[int]HeartbeatEntry
	if err := server.Call("ServerRPC.GetMembershipList", &clientID, &membershipList); err == nil {
		fmt.Printf("Node %d received membership list: %v\n", clientID, membershipList)
		// Add entries to our local table
		client.Mutex.Lock()
		for id, entry := range membershipList {
			client.Table.Entries[id] = entry
		}
		client.Mutex.Unlock()
	}


	// Connect to peers
	for _, peerAddr := range reply.Peers {
		peer, err := rpc.Dial("tcp", peerAddr)
		if err == nil {
			client.Peers = append(client.Peers, peer)
			fmt.Printf("Node %d connected to peer at %s\n", clientID, peerAddr)
		} else {
			fmt.Printf("Node %d failed to connect to peer at %s\n", clientID, peerAddr)
		}
	}

	// Start heartbeat and gossip
	go client.StartHeartbeat(1 * time.Second) // X = 2s
	go client.SendGossip(2 * time.Second) // Y = 5s
	go client.CheckNodeFailures(2 * time.Second) // Run failure check every 2 seconds

	select {} // Keep running
} 