# Gorums reconfigurable storage

This project contains a simple replicated storage system, implemented using [Gorums](https://github.com/relab/gorums/blob/master/doc/user-guide.md).
The system in this module is reconfigurable, meaning the set of servers used to store values can be changes.


## About

This system extends the [non-reconfigurable version](../storage/).

### Representing configurations
For simplicity we only use majority quorums.
We use a somewhat simplified variant to represent servers in a configuration as strings:
* A string `"1:4"` represents the servers stored at index 1,2, and 3 on the client. When using multiple clients, make sure server addresses are submitted in the same order.
* A string `"0,2,3"` represents the servers stored at index 0,2, and 3 on the client.

*Method `c.parseConfiguration()` in `client.go` can be used to convert these string into a `Configuration` from the Gorums library, on which quorum call can be invoked.*

Meta-information about configurations is represented as protobuf `Config` message:
```protobuf
message Config {
  bool Started = 1;
  string Adds = 2;
  google.protobuf.Timestamp Time = 3;
}
```

* `Adds` represents the servers in the above string notation.
* `Time` is a timestamp to distinguish old and new configurations.
* `Started` indicates whether a reconfiguration towards this configuration was completed.

**Timestamps are assumed to be unique.**

### Configuration handling server side

In this system, the server does not handle RPCs differently depending on the configuration on which they are invoked. 
Indeed, the servers does not receive this information.
However, the server does store some Configuration information and returns this to the client on RPCs.
`read`, `write`, and `list` RPCs now also return a list of configurations.
The Quorum functions in `qspec.go` have been updated to combine and de-dublicate the lists received from individual RPCs.

The Gorums server includes a new Quorum Call `WriteConfigQC` this can be used to inform the servers about a new configuration.

## Tasks

### #1 Configuration handling server side


Implement the RPC handler used to handle a `WriteConfigQC`.

Add the new configuration to the `s.configs`. However, consider the following:

**If a server knows a started configuration all earlier configurations (lower timestamp) can be removed.**


*Complete the following stub in `server.go`:*
```go
func (s *storageServer) WriteConfig(req *proto.Config) (*proto.WriteResponse, error) {
	s.logger.Printf("Config '%s', started: '%t'\n", req.GetAdds(), req.GetStarted())
	s.mut.Lock()
	defer s.mut.Unlock()

	// add new configuration, consider if it is started
	// consider if it is older than a started configuration

	return &proto.WriteResponse{New: true, MConfigs s.configs}, nil
}
```

### #2 Implement a reconfiguration procedure

Implement a simple reconfiguration procedure that 
* informs the old configuration about the new configuration,
* gets a list of keys from the old configuration,
* reads values from the old configuration and writes them to the new configuration,
* starts the new configuration, by informing the servers that the new configuration is started.

You can use the folloing methods on the client, to read, write, list and writeConfig in different configurations:
```go
func (client) readQC(key string, cfg *proto.Configuration) *proto.ReadResponse
func (client) writeQC(key, value string, cfg *proto.Configuration) *proto.WriteResponse
func (client) listQC(cfg *proto.Configuration) *proto.ListResponse
func (client) writeConfigQC(conf *proto.Config, cfg *proto.Configuration) *proto.WriteResponse
```

*Complete the following stub in `client.go`:*
```go
func (c *client) reconf(newAdds string) {

	goalProtoConf := &proto.Config{Adds: newAdds, Started: false, Time: timestamppb.Now()}
	// create a Configuration used for quorum calls.
	goalCfg := c.parseConfiguration(newAdds)

    // TODO: perform reconfiguration

	// update the default configuration used by the client
	c.cfg = goalCfg
    c.pcfg = goalProtoConf
}
```

### #3 Perform writes during a reconfiguration

When performing a write on an old configuration, the configuration should get updated and the write should be performed on the new configuration as well.
To be able to do a write, while a reconfiguration (state tranfer) is ongoing, we need to write both to the old and new configuration.

Use the list of configurations in `proto.WriteResponse.Config` to perform a write on one configuration and all its successors.
Start with the default configuration of the client `c.cfg`.

*Complete the following stub in `client.go`:*
```go
func (c *client) write(key, value string) *proto.WriteResponse {

	// TODO perform the write on c.cfg and all its successors

	return &proto.WriteResponse{New: false}
}
```

* Try first without looking how the same is done for the read.
* Does your implementation handle the case where configuration C1 references C2, which again references C3?

*Optional: If you find a started configuration, you can skip older configurations and also update the default configuration of the client.*

Tip: The field `pcfg` on the client should contain the metadata including timestamp of the current configuration.

### #4 Update the reconfiguration to handle concurrent reconfigurations

Update the reconfiguration method you created in #2 to be able to handle concurrent reconfigurations, similar to the write operation in #3.

Tip: Use the provided functions already accessing multiple configurations:
```go
func (c *client) writeConfig(target *proto.Config) *proto.WriteResponse
func (c *client) list() *proto.ListResponse
func (c *client) read(key string) *proto.ReadResponse
```
