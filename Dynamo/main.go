package main

import (
	"flag"
	"fmt"
	"log"
	//"net/rpc"
	"os"
)

const (
	basePort               = 8000
	membershipServerAddr   = "localhost:9000"
)

func main() {
	// Provide only a unique node ID.
	id := flag.Int("id", 0, "Unique node ID (integer, e.g., 1, 2, 3, ...)")
	flag.Parse()
	if *id == 0 {
		fmt.Println("Usage: go run main.go -id <nodeID>")
		os.Exit(1)
	}

	// Automatically assign an address using the node ID.
	address := fmt.Sprintf("localhost:%d", basePort+*id)
	nodeID := fmt.Sprintf("node%d", *id)
	log.Printf("Starting %s at %s", nodeID, address)

	node := NewNode(nodeID, address)

	// Register with the central membership server.
	err := node.RegisterWithCentral()
	if err != nil {
		log.Fatalf("Error registering with central membership server: %v", err)
	}

	// Retrieve the initial list of active nodes.
	err = node.FetchMembersFromCentral()
	if err != nil {
		log.Printf("Warning: Could not fetch initial members: %v", err)
	}

	// Start the RPC server for handling gossip messages.
	go node.StartRPCServer()

	// Start the heartbeat updater.
	go node.updateHeartbeat()

	// Start the periodic gossip loop.
	node.StartGossip()

	// Block forever.
	select {}
}