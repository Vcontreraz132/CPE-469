// node.go
package main

import (
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"sync"
	"time"
	//"encoding/json"
	"errors"
	"fmt"
)

// Node represents a single node in the system.
type Node struct {
	ID         string
	Address    string
	Peers      map[string]struct{}   // Known peer addresses.
	Membership map[string]Member     // Local membership table.
	// Store *DataStore
	DataStore map[string]ValueRecord
	mutex      sync.Mutex
	Ring *Ring
	OperationLog []string
	LogMutex sync.Mutex
}

// NewNode creates a new Node instance.
func NewNode(id, address string) *Node {
	node := &Node{
		ID:         id,
		Address:    address,
		Peers:      make(map[string]struct{}),
		Membership: make(map[string]Member),
		// Store: NewDataStore(),
		DataStore: make(map[string]ValueRecord),
	}
	// Add self to membership table.
	node.Membership[id] = Member{
		ID:       id,
		Address:  address,
		LastSeen: time.Now(),
	}
	node.Ring = BuildRingFromMembership(node.Membership) // add node to ring upon startup
	return node
}

// RegisterWithCentral contacts the central membership server to register.
func (n *Node) RegisterWithCentral() error {
	client, err := rpc.Dial("tcp", membershipServerAddr)
	if err != nil {
		return err
	}
	defer client.Close()

	args := &RegisterArgs{
		Member: Member{
			ID:       n.ID,
			Address:  n.Address,
			LastSeen: time.Now(),
		},
	}
	var reply RegisterReply
	err = client.Call("MembershipServer.Register", args, &reply)
	if err != nil {
		return err
	}
	return nil
}

// FetchMembersFromCentral retrieves the current member list.
func (n *Node) FetchMembersFromCentral() error {
	client, err := rpc.Dial("tcp", membershipServerAddr)
	if err != nil {
		return err
	}
	defer client.Close()

	args := &GetMembersArgs{}
	var reply GetMembersReply
	err = client.Call("MembershipServer.GetMembers", args, &reply)
	if err != nil {
		return err
	}
	n.mutex.Lock()
	defer n.mutex.Unlock()
	for _, member := range reply.Members {
		if member.ID != n.ID {
			n.Membership[member.ID] = member
			n.Peers[member.Address] = struct{}{}
		}
	}
	return nil
}

// --- RPC-based Gossip Service ---

// GossipService provides RPC methods for gossip.
type GossipService struct {
	Node *Node
}

// GossipArgs is the argument for the Gossip RPC.
type GossipArgs struct {
	SenderID   string
	Membership map[string]Member
}

// GossipReply is empty.
type GossipReply struct{}

// Gossip receives a gossip message and merges the membership table.
func (gs *GossipService) Gossip(args *GossipArgs, reply *GossipReply) error {
	gs.Node.mergeMembership(args.Membership)
	gs.Node.mutex.Lock()
	defer gs.Node.mutex.Unlock()
	// Optionally, add the sender as a peer.
	if sender, ok := args.Membership[args.SenderID]; ok {
		gs.Node.Peers[sender.Address] = struct{}{}
	}
	log.Printf("Received gossip from %s", args.SenderID)
	return nil
}

// StartRPCServer starts an RPC server for this node.
func (n *Node) StartRPCServer() {
	gossipService := &GossipService{Node: n}
	DataService := &DataService{Node: n}
	rpc.Register(gossipService)
	rpc.Register(DataService)
	listener, err := net.Listen("tcp", n.Address)
	if err != nil {
		log.Fatalf("Error starting RPC server on %s: %v", n.Address, err)
	}
	log.Printf("RPC server listening on %s", n.Address)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting RPC connection: %v", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}

const GossipInterval = 2 * time.Second

// StartGossip begins the periodic gossiping.
func (n *Node) StartGossip() {
	ticker := time.NewTicker(GossipInterval)
	go func() {
		for range ticker.C {
			n.gossip()
		}
	}()
}

// gossip picks a random peer and sends the membership snapshot via RPC.
func (n *Node) gossip() {
	n.mutex.Lock()
	peers := make([]string, 0, len(n.Peers))
	for p := range n.Peers {
		peers = append(peers, p)
	}
	n.mutex.Unlock()

	if len(peers) == 0 {
		return
	}
	target := peers[rand.Intn(len(peers))]

	args := &GossipArgs{
		SenderID:   n.ID,
		Membership: n.getMembershipSnapshot(),
	}
	var reply GossipReply
	client, err := rpc.Dial("tcp", target)
	if err != nil {
		log.Printf("Error dialing peer %s: %v", target, err)
		return
	}
	defer client.Close()
	err = client.Call("GossipService.Gossip", args, &reply)
	if err != nil {
		log.Printf("Error calling gossip on %s: %v", target, err)
		return
	}
	log.Printf("Gossip sent to %s", target)
}

// getMembershipSnapshot returns a copy of the membership table.
func (n *Node) getMembershipSnapshot() map[string]Member {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	snapshot := make(map[string]Member)
	for k, v := range n.Membership {
		snapshot[k] = v
	}
	return snapshot
}

// mergeMembership merges incoming membership data.
func (n *Node) mergeMembership(incoming map[string]Member) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	changed := false
	for id, member := range incoming {
		existing, exists := n.Membership[id]
		if !exists || member.LastSeen.After(existing.LastSeen) {
			n.Membership[id] = member
			n.Peers[member.Address] = struct{}{}
			
			// if !exists {
			// n.Ring.build(member)
			// fmt.Println("Member added")
			changed = true
			// }

			log.Printf("New membership: %s at %s", id, member.Address)
		}
	}
	if changed {
		n.Ring = BuildRingFromMembership(n.Membership)
		log.Printf("Ring updated: %s", n.Ring.String())
	}
}

// updateHeartbeat periodically updates this node's own timestamp.
func (n *Node) updateHeartbeat() {
	for {
		time.Sleep(GossipInterval)
		n.mutex.Lock()
		member := n.Membership[n.ID]
		member.LastSeen = time.Now()
		n.Membership[n.ID] = member
		n.mutex.Unlock()
	}
}

// -- consensus / data ops --

// create vector clock to map node ids to counter
type VectorClock map[string]int

// create deep copy of vector clock so that you can modify vector clock while keeping a backup
func (vc VectorClock) Copy()VectorClock {
	newVc := make(VectorClock)
	for k, v := range vc {
		newVc[k] = v
	}
	return newVc
}

// return sum of all counters to the latest version of vector clock
func sumClock(vc VectorClock) int{
	sum := 0
	for _, v := range vc {
		sum += v
	}
	return sum
}

// store value with vector clock
type ValueRecord struct {
	Value string
	Clock VectorClock
}

// --- Consensus / Data Operations ---

// DoPut implements a consensus-based write for a given key and value.
// It performs a local update and then contacts a quorum of peers.
func (n *Node) DoPut(key, value string) error {
	// Local update: update the vector clock.
	n.mutex.Lock()
	var newClock VectorClock
	if record, exists := n.DataStore[key]; exists {
		newClock = record.Clock.Copy()
	} else {
		newClock = make(VectorClock)
	}
	newClock[n.ID]++
	newRecord := ValueRecord{
		Value: value,
		Clock: newClock,
	}
	n.DataStore[key] = newRecord
	n.mutex.Unlock()
	log.Printf("DoPut: Node %s locally updated key %s to value %s with clock %v", n.ID, key, value, newClock)

	n.LogOperation("PUT_LOCAL", fmt.Sprintf("Key %s updated to '%s' with clock %v", key, value, newClock))

	// Select quorum peers (for replication factor N=3, select 2 peers).
	quorumPeers := n.selectQuorumPeers()
	ackCount := 1 // local update counts as one ack.
	for _, peer := range quorumPeers {
		args := &PutArgs{
			Key:   key,
			Value: value,
			Clock: newClock,
		}
		var reply PutReply
		client, err := rpc.Dial("tcp", peer)
		if err != nil {
			log.Printf("DoPut: error dialing peer %s: %v", peer, err)
			n.LogOperation("PUT_RPC_ERROR", fmt.Sprintf("Error dialing peer %s: %v", peer, err))
			continue
		}
		err = client.Call("DataService.PutRPC", args, &reply)
		client.Close()
		if err != nil {
			log.Printf("DoPut: error calling PutRPC on %s: %v", peer, err)
			n.LogOperation("PUT_RPC_ERROR", fmt.Sprintf("Error calling PutRPC on %s: %v", peer, err))
			continue
		}
		if reply.Success {
			ackCount++
			n.LogOperation("PUT_RPC_SUCCESS", fmt.Sprintf("Peer %s acknowledged put for key %s", peer, key))
		}
	}
	W := 2 // Write quorum requirement.
	if ackCount >= W {
		log.Printf("DoPut: Quorum achieved for key %s (%d acks)", key, ackCount)
		n.LogOperation("PUT_QUORUM", fmt.Sprintf("Quorum achieved for key %s (%d acks)", key, ackCount))
		return nil
	}
	return errors.New("DoPut: consensus failed, insufficient acknowledgements")
}

// DoGet implements a quorum-based read for a given key.
// It queries a quorum of nodes and returns the value with the highest clock.
func (n *Node) DoGet(key string) (ValueRecord, error) {
	var responses []ValueRecord
	// Get local value (if exists).
	n.mutex.Lock()
	if record, exists := n.DataStore[key]; exists {
		responses = append(responses, record)
		n.LogOperation("GET_LOCAL", fmt.Sprintf("Found key %s locally with clock %v", key, record.Clock))
	}
	n.mutex.Unlock()

	// Query quorum peers.
	quorumPeers := n.selectQuorumPeers()
	for _, peer := range quorumPeers {
		args := &GetArgs{Key: key}
		var reply GetReply
		client, err := rpc.Dial("tcp", peer)
		if err != nil {
			log.Printf("DoGet: error dialing peer %s: %v", peer, err)
			n.LogOperation("GET_RPC_ERROR", fmt.Sprintf("Error dialing peer %s: %v", peer, err))
			continue
		}
		err = client.Call("DataService.GetRPC", args, &reply)
		client.Close()
		if err != nil {
			log.Printf("DoGet: error calling GetRPC on %s: %v", peer, err)
			n.LogOperation("GET_RPC_ERROR", fmt.Sprintf("Error calling GetRPC on %s: %v", peer, err))
			continue
		}
		responses = append(responses, reply.Record)
		n.LogOperation("GET_RPC_SUCCESS", fmt.Sprintf("Peer %s returned value for key %s with clock %v", peer, key, reply.Record.Clock))
	}
	if len(responses) == 0 {
		return ValueRecord{}, errors.New("DoGet: key not found in quorum")
	}
	// For simplicity, select the record with the highest sum of clock counters.
	latest := responses[0]
	for _, rec := range responses[1:] {
		if sumClock(rec.Clock) > sumClock(latest.Clock) {
			latest = rec
		}
	}
	log.Printf("DoGet: Returning key %s with value %s and clock %v", key, latest.Value, latest.Clock)
	n.LogOperation("GET_QUORUM", fmt.Sprintf("Returning key %s with value '%s' and clock %v", key, latest.Value, latest.Clock))
	return latest, nil
}

// selectQuorumPeers selects up to 2 random peer addresses from the membership.
func (n *Node) selectQuorumPeers() []string {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	var peers []string
	for addr := range n.Peers {
		peers = append(peers, addr)
	}
	if len(peers) > 2 {
		rand.Shuffle(len(peers), func(i, j int) { peers[i], peers[j] = peers[j], peers[i] })
		peers = peers[:2]
	}
	return peers
}



// operation log with timestamp
func (n *Node) LogOperation(opType, details string) {
	entry := fmt.Sprintf("%s - %s: %s", time.Now().Format(time.RFC3339), opType, details)
	n.LogMutex.Lock()
	n.OperationLog = append(n.OperationLog, entry)
	n.LogMutex.Unlock()
	// Also output to standard log for real-time observation.
	log.Println(entry)
}


const heartbeatTimeout = 20 * time.Second // adjust as needed

// cleanupStaleMembers periodically removes members whose heartbeat is stale.
func (n *Node) cleanupStaleMembers() {
    for {
        time.Sleep(5 * time.Second) // check every 5 seconds
        n.mutex.Lock()
        for id, member := range n.Membership {
            if id == n.ID {
                continue // skip self
            }
            if time.Since(member.LastSeen) > heartbeatTimeout {
                log.Printf("Removing stale member %s (last seen %v ago)", id, time.Since(member.LastSeen))
                delete(n.Membership, id)
                delete(n.Peers, member.Address)
				n.Ring.RemoveMember(id)
				n.Ring.printNeighbors(id)
            }
        }
        n.mutex.Unlock()
    }
}
