package main

import (
	"time"
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
	ID      int
	State   status        //Describes the status
	Timeout time.Duration //Timeout counter
	Term    int           //Tracks the term
}
