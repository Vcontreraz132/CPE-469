// data.go
package main

import (
	"errors"
	"log"
)

// DataService provides RPC methods for data operations.
type DataService struct {
	Node *Node
}

// PutArgs is used for PutRPC.
type PutArgs struct {
	Key   string
	Value string
	Clock VectorClock
}

// PutReply indicates success or failure.
type PutReply struct {
	Success bool
}

// GetArgs is used for GetRPC.
type GetArgs struct {
	Key string
}

// GetReply returns the stored value record.
type GetReply struct {
	Record ValueRecord
}

// PutRPC stores (or updates) the key–value pair in the local DataStore.
func (ds *DataService) PutRPC(args *PutArgs, reply *PutReply) error {
	ds.Node.mutex.Lock()
	defer ds.Node.mutex.Unlock()
	// In a full implementation, one would merge vector clocks here.
	ds.Node.DataStore[args.Key] = ValueRecord{
		Value: args.Value,
		Clock: args.Clock,
	}
	log.Printf("DataService: Node %s updated key %s to value %s with clock %v", ds.Node.ID, args.Key, args.Value, args.Clock)
	reply.Success = true
	return nil
}

// GetRPC returns the value record for a given key.
func (ds *DataService) GetRPC(args *GetArgs, reply *GetReply) error {
	ds.Node.mutex.Lock()
	defer ds.Node.mutex.Unlock()
	record, exists := ds.Node.DataStore[args.Key]
	if !exists {
		return errors.New("DataService: key not found")
	}
	reply.Record = record
	return nil
}

// ForwardPutRPC is an RPC method that forwards a put request to the responsible node.
func (ds *DataService) ForwardPutRPC(args *PutArgs, reply *PutReply) error {
	// Check for an existing value.
	ds.Node.mutex.Lock()
	existing, exists := ds.Node.DataStore[args.Key]
	ds.Node.mutex.Unlock()

	if exists {
		log.Printf("Node %s received forwarded put for key %s. Changing previous value '%s' to '%s'", ds.Node.ID, args.Key, existing.Value, args.Value)
	} else {
		log.Printf("Node %s received forwarded put for key %s. No previous value, setting to '%s'", ds.Node.ID, args.Key, args.Value)
	}

	// Process the put on the responsible node.
	err := ds.Node.DoPut(args.Key, args.Value)
	if err != nil {
		reply.Success = false
		return err
	}
	reply.Success = true
	return nil
}

// ForwardGetRPC forwards a get request to the responsible node.
func (ds *DataService) ForwardGetRPC(args *GetArgs, reply *GetReply) error {
	log.Printf("Node %s received forwarded get request for key %s", ds.Node.ID, args.Key)

	record, err := ds.Node.DoGet(args.Key)
	if err != nil {
		return err
	}
	reply.Record = record
	return nil
}
