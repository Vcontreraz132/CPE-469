// ring.go
package main

import (
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
	"sync"
	//"time"
)

// RingNode represents an entry in the hash ring for a given member.
type RingNode struct {
	Member Member // Uses the Member struct defined in struct.go
	Hash   uint32 // The computed hash of the member's ID.
}

// Ring represents the circular hash ring.
type Ring struct {
	nodes []RingNode
	mutex sync.Mutex
}

// NewRing creates a new, empty ring.
func NewRing() *Ring {
	return &Ring{
		nodes: make([]RingNode, 0),
	}
}

// hashValue computes a uint32 hash for a given key using MD5.
// Here we use the first 4 bytes of the MD5 hash.
func hashValue(key string) uint32 {
	sum := md5.Sum([]byte(key))
	return binary.BigEndian.Uint32(sum[:4])
}

// BuildRingFromMembership builds the ring using the provided membership table.
// It does not create new members—it simply uses the members already in the map.
func BuildRingFromMembership(membership map[string]Member) *Ring {
	ring := NewRing()
	for _, m := range membership {
		hash := hashValue(m.ID)
		rn := RingNode{
			Member: m,
			Hash:   hash,
		}
		ring.nodes = append(ring.nodes, rn)
	}
	// Sort the ring by hash value.
	sort.Slice(ring.nodes, func(i, j int) bool {
		return ring.nodes[i].Hash < ring.nodes[j].Hash
	})
	return ring
}

// AddMember adds a new member (node) to the ring and prints its neighbors.
func (r *Ring) AddMember(m Member) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	hash := hashValue(m.ID)
	rn := RingNode{
		Member: m,
		Hash:   hash,
	}
	r.nodes = append(r.nodes, rn)
	// Sort the ring by hash value.
	sort.Slice(r.nodes, func(i, j int) bool {
		return r.nodes[i].Hash < r.nodes[j].Hash
	})
	fmt.Printf("Added member %s with hash %d to ring\n", m.ID, hash)
	r.printNeighbors(m.ID)
}

// RemoveMember removes a member from the ring by its ID.
func (r *Ring) RemoveMember(memberID string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for i, rn := range r.nodes {
		if rn.Member.ID == memberID {
			r.nodes = append(r.nodes[:i], r.nodes[i+1:]...)
			fmt.Printf("Removed member %s from ring\n", memberID)
			return
		}
	}
	fmt.Printf("Member %s not found in ring\n", memberID)
}

// GetRingNodes returns a copy of the current list of ring nodes.
func (r *Ring) GetRingNodes() []RingNode {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	copyNodes := make([]RingNode, len(r.nodes))
	copy(copyNodes, r.nodes)
	return copyNodes
}

// String returns a string representation of the ring for debugging.
func (r *Ring) String() string {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	s := "Ring: "
	for _, rn := range r.nodes {
		s += fmt.Sprintf("%s(%d) -> ", rn.Member.ID, rn.Hash)
	}
	return s
}

// GetSuccessor returns the immediate successor of the given member.
func (r *Ring) GetSuccessor(memberID string) (RingNode, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	index := -1
	for i, rn := range r.nodes {
		if rn.Member.ID == memberID {
			index = i
			break
		}
	}
	if index == -1 {
		return RingNode{}, errors.New("member not found")
	}
	successorIndex := (index + 1) % len(r.nodes)
	return r.nodes[successorIndex], nil
}

// GetPredecessor returns the immediate predecessor of the given member.
func (r *Ring) GetPredecessor(memberID string) (RingNode, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	index := -1
	for i, rn := range r.nodes {
		if rn.Member.ID == memberID {
			index = i
			break
		}
	}
	if index == -1 {
		return RingNode{}, errors.New("member not found")
	}
	predecessorIndex := (index - 1 + len(r.nodes)) % len(r.nodes)
	return r.nodes[predecessorIndex], nil
}

// GetResponsibleNode returns the ring node responsible for the given key.
func (r *Ring) GetResponsibleNode(key string) (RingNode, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if len(r.nodes) == 0 {
		return RingNode{}, errors.New("Ring is empty")
	}

	keyHash := hashValue(key)
	// Binary search for the first node with hash >= keyHash.
	idx := sort.Search(len(r.nodes), func(i int) bool {
		return r.nodes[i].Hash >= keyHash
	})

	// If not found, wrap around to the first node.
	if idx == len(r.nodes) {
		idx = 0
	}

	return r.nodes[idx], nil
}

// printNeighbors prints the predecessor and successor of the given member.
func (r *Ring) printNeighbors(memberID string) {
	succ, errSucc := r.GetSuccessor(memberID)
	pred, errPred := r.GetPredecessor(memberID)
	if errSucc != nil || errPred != nil {
		fmt.Printf("Could not find neighbors for member %s\n", memberID)
		return
	}
	fmt.Printf("Member %s -> Predecessor: %s (%d), Successor: %s (%d)\n",
		memberID, pred.Member.ID, pred.Hash, succ.Member.ID, succ.Hash)
}
