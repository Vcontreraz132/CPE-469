// file contains node behavior
package main

import (
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"sync"
	"time"
)

// convert status to string during a print //
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

// start RPC server for nodes
func (n *Node) start() error {
	server := rpc.NewServer() // create RPC server for each node
	err := server.Register(n) // attempt to register node
	if err != nil {
		fmt.Printf("ERROR: Node %d failed to register RPC server\n", n.ID)
	}

	listener, err := net.Listen("tcp", n.Address) // listen for requests on node port

	if err != nil {
		fmt.Println("ERROR:", err)
		return nil
	}

	fmt.Printf("Node %d started RPC server on %s\n", n.ID, n.Address)

	// branch and handle client requests
	go func() {
		for {
			connection, err := listener.Accept()

			if err != nil {
				fmt.Println("ERROR:", err)
				continue
			}

			go server.ServeConn(connection)
		}
	}()
	return nil
}

func (n *Node) election() {
	for {
		timeout := time.Duration(2+rand.Intn(10)) * time.Second // random timeout between 1 and 10 seconds
		select {
		case <-time.After(timeout):
			n.mutex.Lock()
			if n.State == follower {
				fmt.Printf("Node %d has timed out. Starting election\n", n.ID)
				n.mutex.Unlock()
				n.startElection()
			} else {
				n.mutex.Unlock()
			}
		case <-n.resetElection:
			//valid heartbeat was received
			//fmt.Printf("Node %d election timer reset\n", n.ID)
		}
	}
}

func (n *Node) startElection() {
	n.mutex.Lock()
	n.State = candidate // node becomes candidate
	n.CurrentTerm++     // increment term count
	term := n.CurrentTerm
	voteCount := 1 // candidate votes for itself
	n.VotedFor = n.ID
	n.mutex.Unlock()
	fmt.Printf("Node %d has become a candidate for term %d\n", n.ID, term)
	fmt.Println("Election starting")

	var wg sync.WaitGroup
	var voteM sync.Mutex

	// send vote requests to peers
	for _, peers := range n.PeerAddress {

		if peers == n.Address { // if peer is this node, skip
			continue
		}

		wg.Add(1) // increment wait count for each node

		go func(peerAddr string) {
			defer wg.Done()
			client, err := rpc.Dial("tcp", peerAddr) // make connection with peer node
			if err != nil {
				fmt.Printf("ERROR: startElection() RPC.dial failed\n")
			}
			defer client.Close()

			request := VoteRequest{ // populate VoteRequest struct
				Term:        term,
				CandidateID: n.ID,
			}

			var reply VoteReply                                    // create a reply
			err = client.Call("Node.RequestVote", request, &reply) // tell peer to execute RequestVote funct
			if err == nil && reply.VoteGranted {                   // if reply came back
				voteM.Lock()
				voteCount++ // increment candidate vote count
				voteM.Unlock()
			}
		}(peers)
	}

	// wait for votes to come in
	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneCh)
	}()
	select {
	case <-doneCh:
	case <-time.After(2 * time.Second):
	}

	n.mutex.Lock()

	if n.State != candidate || n.CurrentTerm != term {
		n.mutex.Unlock()
		return
	}

	if voteCount > len(n.PeerAddress)/2 { // if candidate gets more than half the votes
		fmt.Printf("Node %d wins with %d votes\n", n.ID, voteCount)
		n.State = leader
		n.mutex.Unlock()
		go n.leadership()
	} else {
		fmt.Printf("Node %d lost the election with %d votes\n", n.ID, voteCount)
		n.State = follower
		n.VotedFor = -1
		n.mutex.Unlock()
	}
}

func (n *Node) RequestVote(req VoteRequest, reply *VoteReply) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	// If the candidate's term is older, reject the vote.
	if req.Term < n.CurrentTerm {
		reply.Term = n.CurrentTerm
		reply.VoteGranted = false
		fmt.Printf("Node %d rejects vote for candidate %d (term too old, current term %d)\n", n.ID, req.CandidateID, n.CurrentTerm)
		return nil
	}

	// If the term is higher than our current term, update our term and convert to Follower.
	if req.Term > n.CurrentTerm {
		n.CurrentTerm = req.Term
		n.State = follower
		n.VotedFor = -1
	}

	// if node has not yet voted, cast vote
	if n.VotedFor < 0 || n.VotedFor == req.CandidateID {
		n.VotedFor = req.CandidateID
		reply.VoteGranted = true
		fmt.Printf("Node %d has voted for candidate %d\n", n.ID, req.CandidateID)
	} else { // node has already voted
		reply.VoteGranted = false
		fmt.Printf("Node %d has already voted this term\n")
	}
	reply.Term = n.CurrentTerm
	return nil
}

func (n *Node) leadership() {
	fmt.Printf("Node %d is acting leader for term %d\n", n.ID, n.CurrentTerm)
	timer := time.NewTicker(1 * time.Second)
	defer timer.Stop()

	// Simulate leader failure after 5 seconds //
	go func() {
		time.Sleep(5 * time.Second)
		n.Stop() //Simulate the leader node crashing
	}()

	for {
		n.mutex.Lock()

		if n.State != leader {
			n.mutex.Unlock()
			fmt.Printf("Node %d exiting leader loop\n", n.ID)
			return
		}

		CurrentTerm := n.CurrentTerm
		n.mutex.Unlock()

		// Send heartbeat to peers
		for _, peers := range n.PeerAddress {
			if peers == n.Address {
				continue
			}

			go func(peerAddr string) {
				client, err := rpc.Dial("tcp", peerAddr)
				if err != nil {
					return
				}
				defer client.Close()

				message := HeartbeatSend{
					Term:     CurrentTerm,
					LeaderID: n.ID,
				}

				var reply HeartbeatReply
				client.Call("Node.Heartbeat", message, &reply)
			}(peers)
		}
		<-timer.C
	}
}

func (n *Node) Heartbeat(msg HeartbeatSend, reply *HeartbeatReply) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	fmt.Printf("Node %d received heartbeat from leader %d (term %d)\n", n.ID, msg.LeaderID, msg.Term)

	if msg.Term < n.CurrentTerm {
		reply.Term = n.CurrentTerm
		return nil
	}

	n.CurrentTerm = msg.Term // update nodes term tracker

	n.State = follower // force node to become follower
	n.VotedFor = -1    // reset vote flag

	// Signal to reset our election timer.
	select {
	case n.resetElection <- true:
		fmt.Printf("Node %d resets election timer on heartbeat from leader %d\n", n.ID, msg.LeaderID)
	default:
		// If channel is full, that’s okay.
	}

	reply.Term = n.CurrentTerm
	return nil
}

func (n *Node) Stop() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.State = offline
	fmt.Printf("Node %d has gone offline (Leader failed)\n", n.ID)
}
