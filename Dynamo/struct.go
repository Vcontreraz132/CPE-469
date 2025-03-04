package main

import "time"

// Member holds a node's membership information.
type Member struct {
	ID       string 
	Address  string
	LastSeen time.Time
}

// Definitions for central membership RPC arguments and replies.

type RegisterArgs struct {
	Member Member
}

type RegisterReply struct{}

type UnregisterArgs struct {
	ID string
}

type UnregisterReply struct{}

type GetMembersArgs struct{}

type GetMembersReply struct {
	Members map[string]Member
}