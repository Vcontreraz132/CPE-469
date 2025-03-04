package main

import (
	"fmt"
	"log"
	"net/rpc"
	"time"
)

// ConsensusPut sends a put request to a replication set and waits for a write quorum (W).
func (n *Node) ConsensusPut(key, value string) error {
	// For this simulation, we set:
	N := 3 // replication factor
	W := 2 // write quorum

	// Gather a list of known node addresses.
	n.mutex.Lock()
	var nodes []string
	for _, member := range n.Membership {
		nodes = append(nodes, member.Address)
	}
	n.mutex.Unlock()

	if len(nodes) < N {
		return fmt.Errorf("not enough nodes to meet replication factor (have %d, need %d)", len(nodes), N)
	}

	// For simplicity, pick the first N nodes.
	replicationNodes := nodes[:N]

	// Create an initial vector clock for the write.
	vc := make(map[string]int)
	vc[n.ID] = 1

	successCount := 0
	for _, addr := range replicationNodes {
		client, err := rpc.Dial("tcp", addr)
		if err != nil {
			log.Printf("Error dialing node at %s: %v", addr, err)
			continue
		}
		args := &PutArgs{
			Key:    key,
			Value:  value,
			VClock: vc,
		}
		var reply PutReply
		err = client.Call("DataService.Put", args, &reply)
		client.Close()
		if err != nil {
			log.Printf("Error calling Put on %s: %v", addr, err)
		} else {
			successCount++
		}
	}

	if successCount >= W {
		log.Printf("ConsensusPut succeeded for key '%s' (%d/%d nodes)", key, successCount, N)
		return nil
	}
	return fmt.Errorf("ConsensusPut failed: only %d successful responses", successCount)
}

// ConsensusGet queries a replication set and waits for a read quorum (R).
func (n *Node) ConsensusGet(key string) (DataItem, error) {
	// For this simulation, we set:
	N := 3 // replication factor
	R := 2 // read quorum

	// Gather list of known nodes.
	n.mutex.Lock()
	var nodes []string
	for _, member := range n.Membership {
		nodes = append(nodes, member.Address)
	}
	n.mutex.Unlock()

	if len(nodes) < N {
		return DataItem{}, fmt.Errorf("not enough nodes to meet replication factor (have %d, need %d)", len(nodes), N)
	}
	replicationNodes := nodes[:N]

	type result struct {
		item DataItem
		err  error
	}
	resultsChan := make(chan result, len(replicationNodes))
	for _, addr := range replicationNodes {
		go func(address string) {
			client, err := rpc.Dial("tcp", address)
			if err != nil {
				resultsChan <- result{err: err}
				return
			}
			args := &GetArgs{Key: key}
			var reply GetReply
			err = client.Call("DataService.Get", args, &reply)
			client.Close()
			resultsChan <- result{item: reply.Data, err: err}
		}(addr)
	}

	var responses []DataItem
	timeout := time.After(2 * time.Second)
	for i := 0; i < len(replicationNodes); i++ {
		select {
		case res := <-resultsChan:
			if res.err == nil {
				responses = append(responses, res.item)
			}
			if len(responses) >= R {
				// In a real system, you’d merge vector clocks and resolve conflicts.
				// For simplicity, here we choose the item with the highest value for our own node's counter.
				chosen := responses[0]
				if len(responses) > 1 {
					for _, item := range responses[1:] {
						if item.VClock[n.ID] > chosen.VClock[n.ID] {
							chosen = item
						}
					}
				}
				return chosen, nil
			}
		case <-timeout:
			return DataItem{}, fmt.Errorf("ConsensusGet timed out")
		}
	}
	return DataItem{}, fmt.Errorf("ConsensusGet failed: only %d responses", len(responses))
}