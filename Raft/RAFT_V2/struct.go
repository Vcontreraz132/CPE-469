// struct.go
// file contains common struct and var defs
package main

import (
	"sync"
	"fmt"
)

// Enumerates possible node status //
type status int64

const (
	follower status = iota
	candidate
	leader
	offline
)

// Node structure //
type Node struct {
	mutex sync.Mutex
	ID int
	State status //Describes the status
	CurrentTerm int //Tracks the term
	VotedFor int // ID of vote
	VoteCount int // count of candidate votes
	Address string // node's address
	PeerAddress []string // member list
	resetElection chan bool // reset election timer
}

type VoteRequest struct {
	Term int // current term number
	CandidateID int // ID of node running election
}

type VoteReply struct {
	Term int // cast vote for this term
	VoteGranted bool // has the node voted already?
}

type HeartbeatSend struct {
	Term int // current term = valid heartbeat
	LeaderID int // heartbeat came from current leader
}

type HeartbeatReply struct {
	Term int
}

func (n *Node) Print() {
	fmt.Printf("Node %d is %s (term %d)\n", n.ID, n.State, n.CurrentTerm) // print status of node
}
