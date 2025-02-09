package main

// Enumerates possible node status //
type status int64

const (
	follower status = iota
	candidate
	leader
	offline
)

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

// Node structure //
type Node struct {
	ID        int
	State     status //Describes the status
	Heartbeat chan bool
	Term      int //Tracks the term
}
