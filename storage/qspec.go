package main

import (
	"storage/proto"
)

type qspec struct {
	cfgSize int
}

// ReadQCQF is the quorum function for the ReadQC
// ordered quorum call method. The in parameter is the request object
// supplied to the ReadQC method at call time, and may or may not
// be used by the quorum function. If the in parameter is not needed
// you should implement your quorum function with '_ *ReadRequest'.
func (q qspec) ReadQCQF(_ *proto.ReadRequest, replies map[uint32]*proto.ReadResponse) (*proto.ReadResponse, bool) {
	// wait until at least half of the replicas have responded
	if len(replies) <= q.cfgSize/2 {
		return nil, false
	}
	// return the value with the most recent timestamp
	return newestValue(replies), true
}

// WriteQCQF is the quorum function for the WriteQC
// ordered quorum call method. The in parameter is the request object
// supplied to the WriteQC method at call time, and may or may not
// be used by the quorum function. If the in parameter is not needed
// you should implement your quorum function with '_ *WriteRequest'.
func (q qspec) WriteQCQF(in *proto.WriteRequest, replies map[uint32]*proto.WriteResponse) (*proto.WriteResponse, bool) {
	// TODO:
	// * wait for q.cfgSize/2 many replies
	// * if all replies show New, return New: true

	return &proto.WriteResponse{New: false}, true
}

func (q qspec) ListKeysQCQF(in *proto.ListRequest, replies map[uint32]*proto.ListResponse) (*proto.ListResponse, bool) {
	if len(replies) <= q.cfgSize/2 {
		return nil, false
	}
	var keys map[string]bool
	for _, resp := range replies {
		if len(keys) == 0 {
			keys = make(map[string]bool, len(resp.GetKeys()))
		}
		for _, k := range resp.GetKeys() {
			keys[k] = true
		}
	}
	allkeys := make([]string, 0, len(keys))
	for k := range keys {
		allkeys = append(allkeys, k)
	}
	return &proto.ListResponse{Keys: allkeys}, true
}

// newestValue returns the reply that had the most recent timestamp
func newestValue(values map[uint32]*proto.ReadResponse) *proto.ReadResponse {
	if len(values) < 1 {
		return nil
	}
	var newest *proto.ReadResponse
	for _, v := range values {
		if v.GetTime().AsTime().After(newest.GetTime().AsTime()) {
			newest = v
		}
	}
	return newest
}
