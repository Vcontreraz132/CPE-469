package main

import (
	"errors"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"
)

// Member represents a node's information.
type Member struct {
	ID       string
	Address  string
	LastSeen time.Time
}

// MembershipServer maintains the list of active nodes.
type MembershipServer struct {
	mutex   sync.Mutex
	Members map[string]Member
}

// RegisterArgs contains the node information.
type RegisterArgs struct {
	Member Member
}

// RegisterReply is empty.
type RegisterReply struct{}

// UnregisterArgs identifies a node to remove.
type UnregisterArgs struct {
	ID string
}

// UnregisterReply is empty.
type UnregisterReply struct{}

// GetMembersArgs is empty.
type GetMembersArgs struct{}

// GetMembersReply returns the current member list.
type GetMembersReply struct {
	Members map[string]Member
}

// Register a new node.
func (ms *MembershipServer) Register(args *RegisterArgs, reply *RegisterReply) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	ms.Members[args.Member.ID] = args.Member
	log.Printf("Registered node %s at %s", args.Member.ID, args.Member.Address)
	return nil
}

// Unregister removes a node.
func (ms *MembershipServer) Unregister(args *UnregisterArgs, reply *UnregisterReply) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	if _, ok := ms.Members[args.ID]; !ok {
		return errors.New("node not found")
	}
	delete(ms.Members, args.ID)
	log.Printf("Unregistered node %s", args.ID)
	return nil
}

// GetMembers returns the current member list.
func (ms *MembershipServer) GetMembers(args *GetMembersArgs, reply *GetMembersReply) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	copyMembers := make(map[string]Member)
	for k, v := range ms.Members {
		copyMembers[k] = v
	}
	reply.Members = copyMembers
	return nil
}

func main() {
	ms := &MembershipServer{
		Members: make(map[string]Member),
	}
	rpc.Register(ms)
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("Error starting membership server: %v", err)
	}
	log.Println("Membership server listening on :9000")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}