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

// Global variables
var (
	mutex       sync.Mutex
	nodeArray   []Node
	leaderNode  *Node
	leaderFound = false
)

func main() {
	var wg sync.WaitGroup

	// Slice to store nodes //
	numNodes := 8
	nodeArray = make([]Node, numNodes)

	// Initialize Node struct //
	for i := 0; i < numNodes; i++ {
		nodeArray[i] = Node{
			ID:        i,
			State:     follower,
			Heartbeat: make(chan bool),
			Term:      1,
		}
	}

	// Start node process //
	for i := range nodeArray {
		wg.Add(1)
		go nodeProcess(&nodeArray[i], &wg)

	}

	wg.Wait()
	fmt.Println("All nodes have finished execution.")
}

func nodeProcess(node *Node, wg *sync.WaitGroup) {
	defer wg.Done() //Decrease counter when function ends

	fmt.Printf("Node %d is running\n", node.ID)

	// Create a timer that expires upon reaching timeout //
	timeout := time.Duration(rand.Intn(5)+4) * time.Second //Change to 151+150ms later
	timer := time.NewTimer(timeout)

	// Record the first node to time out //
	mutex.Lock()
	if leaderFound == false {
		leaderNode = node
		leaderNode.State = leader
		leaderFound = true //Set global variable

	}
	mutex.Unlock()

	if node.State == leader {
		go sendHeartbeat(leaderNode, 2*time.Second) //Send a heartbeat every interval

	} else if node.State == follower {
		for {
			select {
			case <-node.Heartbeat: //Execute when a heartbeat is detected
				fmt.Printf("Node %d received heartbeat from Leader %d\n", node.ID, leaderNode.ID)

				// Drain timer //
				select {
				case <-timer.C: //Do nothing if timer expired
				default: //Do nothing if a case is not fulfilled
				}

				// Reset timer after receiving a heartbeat //
				newTimeOut := time.Duration(rand.Intn(5)+4) * time.Second
				timer.Reset(newTimeOut)

			case <-timer.C: //Execute if a node times out
				fmt.Printf("Node %d detected Leader failure! Starting election...\n", node.ID)
				leaderFound = false
				return
			}
		}
	}
}

func sendHeartbeat(node *Node, interval time.Duration) {
	timer := time.NewTicker(interval) //Ticker for continous ticks at every interval
	defer timer.Stop()

	for {
		<-timer.C //Prevents execution until the next interval

		// Send a heartbeat to every follower //
		fmt.Printf("\nLeader %d sending heartbeat...\n", node.ID)
		for _, localNode := range nodeArray { //localNode modifies a copy of the struct
			if localNode.ID != node.ID { // Skip leader
				localNode.Heartbeat <- true // Send heartbeat to followers

			}
		}
	}
}
