# Gorums storage

A simple replicated storage system implemented using [Gorums](https://github.com/relab/gorums/blob/master/doc/user-guide.md).

## About

This module implements a simple storage system.
The code includes:
* `server.go` a simple server that store key value pairs. Values are associated with a timestamp.
* `proto\storage.proto` defines the interface of the storage server using protocol buffers.
* `proto` contains stubs for Gorums client and server, code generated from `proto\storage.proto`.
* `qspec.go` contains [quorum functions](https://github.com/relab/gorums/blob/master/doc/user-guide.md#the-quorumspec-interface-with-quorum-functions). These are used by clients to combine RPC responses from multiple servers.
* `client.go` is a convenience wrapper around the RPC client, containing some error handling, ...
* `repl.go` contains a simple read-evaluate-print-loop that can be used to invoke different individual RPCs and Quorum Calls on a set of servers.
* `main.go` contains a wrapper to create a server or client-repl. By default 4 servers and one repl are started.


## Tasks

The following contains a few tasks meant to make you familiar with the code and Gorums Quorum Calls.

### #1 Write function on the server

Write the RPC handler processing write on the server side.
*Complete the following stub in `server.go`:*

```go
// Write writes a new value to storage if it is newer than the old value
func (s *storageServer) Write(req *proto.WriteRequest) (*proto.WriteResponse, error) {
	s.logger.Printf("Write '%s' = '%s'\n", req.GetKey(), req.GetValue())
	s.mut.Lock()
	defer s.mut.Unlock()
    // TODO: 
    // * if req contains a value for a new key, 
    //      store the value and return New: true
    // * else if req contains a value with higher timestamp than the previous,
    //      store the value and return New: true
	return &proto.WriteResponse{New: false}, nil
}
```

### #2 Write the quorum function for writes

A quorum call invokes a write-RPC on multiple processes.
The quorum function receives the individual RPC replies and combines them to a single reply of the same type.
The quorum function is invoked every time an RPC returns and has access to all replies received so far.
The quorum function returns two values, the aggregated or selected reply and a boolean indicating whether 
the client should wait for additional replies, or return the current aggregate, discarding further replies.

*Complete the following stub in `qspec.go`:*

```go
func (q qspec) WriteQCQF(in *proto.WriteRequest, replies map[uint32]*proto.WriteResponse) (*proto.WriteResponse, bool) {
	// TODO:
	// * wait for q.cfgSize/2 many replies
	// * if all replies show New, return New: true

	return &proto.WriteResponse{New: false}, true
}
```

### #3 Experiment with the REPL

Compile the code and run the repl with the 4 default servers:
```
cd reconfstorage/storage
go build
./storage
```

Try the following:
* Invoke a quorum call writing to all servers: `qc write foo bar```
* Invoke an individual rpc on server 0 overwriting the previous value: `rpc 0 write foo spam`
* Invoke an individual rpc on server 1 writing to a new key: `rpc 0 write hello world`
* Invoke a quorum call to read the value `qc read foo` *Do you read the first or second value every time?*
* Change the default configuration on the client to only include servers 1,2,3: `cfg 1-4`
* Invoke a quorum call listing all available keys: `qc list`