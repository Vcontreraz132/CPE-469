/*
 * || Setup Instructions ||
 * 1. F1 -> Click on "Terminal: Create New Terminal" -> Click on "Command Prompt"
 * 2. cd FOLDERNAME
 *
 * || Relevant Terminal Commands ||
 * go run FILENAME.go (Runs file); go build (Build and format files in current directory);
 * gofmt (Format files into standard format);
 *
 * WARNING: Every folder needs a "go.mod" if there isn't one, type "go mod init NAME" in terminal
 *
 * || Instructions ||
 *
 */

package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var timedOutNode *Node
var mutex sync.Mutex

func main() {
	var wg sync.WaitGroup

	// Slice to store nodes //
	numNodes := 8
	nodeArray := make([]Node, numNodes)

	// Initialize Node struct //
	for i := 0; i < numNodes; i++ {
		nodeArray[i] = Node{
			ID:      i,
			State:   follower,
			Timeout: time.Duration(rand.Intn(151)+150) * time.Millisecond,
			Term:    1,
		}
	}

	// Start node process //
	for i := range nodeArray {
		wg.Add(1)
		go nodeProcess(&nodeArray[i], &wg)

	}

	if timedOutNode != nil {
		timedOutNode.State = leader

	}

	wg.Wait()
	fmt.Println("All nodes have finished execution.")
}

// Convert status to string during a print //
func (s status) String() string {
	switch s {
	case follower:
		return "Follower"
	case candidate:
		return "Candidate"
	case leader:
		return "Leader"
	case offline:
		return "Offline"
	default:
		return "Unknown"
	}
}

// Function representing a node process
func nodeProcess(node *Node, wg *sync.WaitGroup) {
	defer wg.Done() // Decrease counter when function ends

	fmt.Printf("Node %d is running\n", node.ID)

	// Decrement timer //
	for node.Timeout > 0 {
		node.Timeout -= time.Millisecond // Decrement by 1ms
		time.Sleep(time.Millisecond)     // Simulate 1ms passing
	}

	// Record the first node to time out //
	mutex.Lock()
	if timedOutNode == nil {
		node.State = leader
		timedOutNode = node

	}
	mutex.Unlock()

	if node.State == leader {
		fmt.Printf("Node: %d -- Hi I'm leader!\n", node.ID)
	}
	if node.State == follower {
		fmt.Printf("Node: %d -- I'm a follower\n", node.ID)
	}
}
