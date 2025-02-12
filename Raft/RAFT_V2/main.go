// main.go
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
	 "time"
 )
 
 func main() {
	rand.Seed(time.Now().UnixNano())
 
	 // Slice to store nodes //
	 numNodes := 8
	 nodeArray := make([]*Node, numNodes)
	 addresses := make([]string, numNodes) // create list for node addresses
	 base := 8000 // base node address

	 // init nodes with an address
	 for i := 0; i < numNodes; i++ {
		addresses[i] = fmt.Sprintf("localhost:%d", base + i) // 8000, 8001 ...
	 }

	 // Initialize Node struct //
	 for i := 0; i < numNodes; i++ {
		 nodeArray[i] = &Node{
			 ID:			i, // current node id
			 State: 		follower, // node starts as follower
			 CurrentTerm:	0, // set current term to 0
			 VotedFor:		-1, // notes havent voted for anyone
			 PeerAddress: addresses, // pop member list
			 Address: addresses[i], // assign node an address
			 resetElection: make(chan bool, 1), // create buffered channel
		 }

		 err := nodeArray[i].start()
		 if err != nil {
			fmt.Printf("Node %d failed to start %v", i, err)
		 }
	 }
 
	 // Start node process //
	 for i := 0; i< numNodes; i++ {
		
		go nodeArray[i].election()
 
	 }
 
	time.Sleep(1 * time.Hour)
	 fmt.Println("All nodes have finished execution.")
 }
