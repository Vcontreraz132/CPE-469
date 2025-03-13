// main.go
package main

import (
	"flag"
	"fmt"
	"log"
	//"net/rpc"
	"os"
	"time"
)

const (
	basePort               = 8000
	membershipServerAddr   = "localhost:9000"
)

func main() {
	// Provide only a unique node ID.
	id := flag.Int("id", 0, "Unique node ID (integer, e.g., 1, 2, 3, ...)")

	testRing := flag.Bool("testRing", false, "Test consistent hash ring functionality")

	flag.Parse()
	
	if *testRing {
		// In test mode, we assume that this node's membership table is already populated.
		// For example, register with the central server and fetch membership.
		if *id == 0 {
			fmt.Println("For testRing mode, please provide a node ID using -id")
			os.Exit(1)
		}
		// Automatically assign an address using the node ID.
		address := fmt.Sprintf("localhost:%d", basePort+*id)
		nodeID := fmt.Sprintf("node%d", *id)
		logFilename := fmt.Sprintf("%s.log", nodeID)
		f, err := os.OpenFile(logFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			fmt.Printf("Error opening log file %s: %v\n", logFilename, err)
			os.Exit(1)
		}
		defer f.Close()
		log.SetOutput(f)

		log.Printf("Starting %s at %s in testRing mode", nodeID, address)

		node := NewNode(nodeID, address)

		// Register with and fetch membership from the central server.
		err = node.RegisterWithCentral()
		if err != nil {
			log.Fatalf("Error registering with central membership server: %v", err)
		}
		err = node.FetchMembersFromCentral()
		if err != nil {
			log.Printf("Warning: Could not fetch initial members: %v", err)
		}

		// Build the ring from the node's current membership.
		ring := BuildRingFromMembership(node.Membership)
		fmt.Println(ring.String())
		// Print neighbors for each member in the membership.
		for memberID := range node.Membership {
			ring.printNeighbors(memberID)
		}
	}

	if *id == 0 {
		fmt.Println("Usage: go run main.go -id <nodeID>")
		os.Exit(1)
	}

	// Automatically assign an address using the node ID.
	address := fmt.Sprintf("localhost:%d", basePort+*id)
	nodeID := fmt.Sprintf("node%d", *id)

	logFilename := fmt.Sprintf("%s.log", nodeID)
	f, err := os.OpenFile(logFilename, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("Error opening log file %s: %v\n", logFilename, err)
		os.Exit(1)
	}
	defer f.Close()
	log.SetOutput(f)

	log.Printf("Starting %s at %s", nodeID, address)

	node := NewNode(nodeID, address)

	// Register with the central membership server.
	err = node.RegisterWithCentral()
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

	// Start the membership cleanup routine.
	go node.cleanupStaleMembers()

	// Start the periodic gossip loop.
	node.StartGossip()

		// For demonstration, you can invoke consensus operations from the coordinator.
	// For example, if this node is node1, after a delay, do a Put and Get.
	// (In practice these would be triggered by client requests.)
	// Uncomment the following block to try a demo operation.
	
		time.Sleep(5 * time.Second)
		err = node.DoPut("exampleKey", "HelloDynamo")
		if err != nil {
			log.Printf("DoPut error: %v", err)
		} else {
			log.Println("DoPut succeeded for key 'exampleKey'")
		}
		time.Sleep(2 * time.Second)
		record, err := node.DoGet("exampleKey")
		if err != nil {
			log.Printf("DoGet error: %v", err)
		} else {
			log.Printf("DoGet returned: %s with clock %v", record.Value, record.Clock)
		}
	

	// Block forever.
	select {}
}