// Code generated by protoc-gen-gorums. DO NOT EDIT.
// versions:
// 	protoc-gen-gorums v0.7.0-devel
// 	protoc            v3.10.0
// source: storage.proto

package proto

import (
	context "context"
	fmt "fmt"
	empty "github.com/golang/protobuf/ptypes/empty"
	gorums "github.com/relab/gorums"
	encoding "google.golang.org/grpc/encoding"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = gorums.EnforceVersion(7 - gorums.MinVersion)
	// Verify that the gorums runtime is sufficiently up-to-date.
	_ = gorums.EnforceVersion(gorums.MaxVersion - 7)
)

// A Configuration represents a static set of nodes on which quorum remote
// procedure calls may be invoked.
type Configuration struct {
	gorums.RawConfiguration
	nodes []*Node
	qspec QuorumSpec
}

// ConfigurationFromRaw returns a new Configuration from the given raw configuration and QuorumSpec.
//
// This function may for example be used to "clone" a configuration but install a different QuorumSpec:
//  cfg1, err := mgr.NewConfiguration(qspec1, opts...)
//  cfg2 := ConfigurationFromRaw(cfg1.RawConfig, qspec2)
func ConfigurationFromRaw(rawCfg gorums.RawConfiguration, qspec QuorumSpec) *Configuration {
	// return an error if the QuorumSpec interface is not empty and no implementation was provided.
	var test interface{} = struct{}{}
	if _, empty := test.(QuorumSpec); !empty && qspec == nil {
		panic("QuorumSpec may not be nil")
	}
	return &Configuration{
		RawConfiguration: rawCfg,
		qspec:            qspec,
	}
}

// Nodes returns a slice of each available node. IDs are returned in the same
// order as they were provided in the creation of the Manager.
//
// NOTE: mutating the returned slice is not supported.
func (c *Configuration) Nodes() []*Node {
	if c.nodes == nil {
		c.nodes = make([]*Node, 0, c.Size())
		for _, n := range c.RawConfiguration {
			c.nodes = append(c.nodes, &Node{n})
		}
	}
	return c.nodes
}

// And returns a NodeListOption that can be used to create a new configuration combining c and d.
func (c Configuration) And(d *Configuration) gorums.NodeListOption {
	return c.RawConfiguration.And(d.RawConfiguration)
}

// Except returns a NodeListOption that can be used to create a new configuration
// from c without the nodes in rm.
func (c Configuration) Except(rm *Configuration) gorums.NodeListOption {
	return c.RawConfiguration.Except(rm.RawConfiguration)
}

func init() {
	if encoding.GetCodec(gorums.ContentSubtype) == nil {
		encoding.RegisterCodec(gorums.NewCodec())
	}
}

// Manager maintains a connection pool of nodes on
// which quorum calls can be performed.
type Manager struct {
	*gorums.RawManager
}

// NewManager returns a new Manager for managing connection to nodes added
// to the manager. This function accepts manager options used to configure
// various aspects of the manager.
func NewManager(opts ...gorums.ManagerOption) (mgr *Manager) {
	mgr = &Manager{}
	mgr.RawManager = gorums.NewRawManager(opts...)
	return mgr
}

// NewConfiguration returns a configuration based on the provided list of nodes (required)
// and an optional quorum specification. The QuorumSpec is necessary for call types that
// must process replies. For configurations only used for unicast or multicast call types,
// a QuorumSpec is not needed. The QuorumSpec interface is also a ConfigOption.
// Nodes can be supplied using WithNodeMap or WithNodeList, or WithNodeIDs.
// A new configuration can also be created from an existing configuration,
// using the And, WithNewNodes, Except, and WithoutNodes methods.
func (m *Manager) NewConfiguration(opts ...gorums.ConfigOption) (c *Configuration, err error) {
	if len(opts) < 1 || len(opts) > 2 {
		return nil, fmt.Errorf("wrong number of options: %d", len(opts))
	}
	c = &Configuration{}
	for _, opt := range opts {
		switch v := opt.(type) {
		case gorums.NodeListOption:
			c.RawConfiguration, err = gorums.NewRawConfiguration(m.RawManager, v)
			if err != nil {
				return nil, err
			}
		case QuorumSpec:
			// Must be last since v may match QuorumSpec if it is interface{}
			c.qspec = v
		default:
			return nil, fmt.Errorf("unknown option type: %v", v)
		}
	}
	// return an error if the QuorumSpec interface is not empty and no implementation was provided.
	var test interface{} = struct{}{}
	if _, empty := test.(QuorumSpec); !empty && c.qspec == nil {
		return nil, fmt.Errorf("missing required QuorumSpec")
	}
	return c, nil
}

// Nodes returns a slice of available nodes on this manager.
// IDs are returned in the order they were added at creation of the manager.
func (m *Manager) Nodes() []*Node {
	gorumsNodes := m.RawManager.Nodes()
	nodes := make([]*Node, 0, len(gorumsNodes))
	for _, n := range gorumsNodes {
		nodes = append(nodes, &Node{n})
	}
	return nodes
}

// Node encapsulates the state of a node on which a remote procedure call
// can be performed.
type Node struct {
	*gorums.RawNode
}

// Reference imports to suppress errors if they are not otherwise used.
var _ empty.Empty

// WriteMulticast is a quorum call invoked on all nodes in configuration c,
// with the same argument in, and returns a combined result.
func (c *Configuration) WriteMulticast(ctx context.Context, in *WriteRequest, opts ...gorums.CallOption) {
	cd := gorums.QuorumCallData{
		Message: in,
		Method:  "storage.Storage.WriteMulticast",
	}

	c.RawConfiguration.Multicast(ctx, cd, opts...)
}

// QuorumSpec is the interface of quorum functions for Storage.
type QuorumSpec interface {
	gorums.ConfigOption

	// ReadQCQF is the quorum function for the ReadQC
	// quorum call method. The in parameter is the request object
	// supplied to the ReadQC method at call time, and may or may not
	// be used by the quorum function. If the in parameter is not needed
	// you should implement your quorum function with '_ *ReadRequest'.
	ReadQCQF(in *ReadRequest, replies map[uint32]*ReadResponse) (*ReadResponse, bool)

	// WriteQCQF is the quorum function for the WriteQC
	// quorum call method. The in parameter is the request object
	// supplied to the WriteQC method at call time, and may or may not
	// be used by the quorum function. If the in parameter is not needed
	// you should implement your quorum function with '_ *WriteRequest'.
	WriteQCQF(in *WriteRequest, replies map[uint32]*WriteResponse) (*WriteResponse, bool)

	// ListKeysQCQF is the quorum function for the ListKeysQC
	// quorum call method. The in parameter is the request object
	// supplied to the ListKeysQC method at call time, and may or may not
	// be used by the quorum function. If the in parameter is not needed
	// you should implement your quorum function with '_ *ListRequest'.
	ListKeysQCQF(in *ListRequest, replies map[uint32]*ListResponse) (*ListResponse, bool)

	// WriteMetaConfQCQF is the quorum function for the WriteMetaConfQC
	// quorum call method. The in parameter is the request object
	// supplied to the WriteMetaConfQC method at call time, and may or may not
	// be used by the quorum function. If the in parameter is not needed
	// you should implement your quorum function with '_ *MetaConfig'.
	WriteMetaConfQCQF(in *MetaConfig, replies map[uint32]*WriteResponse) (*WriteResponse, bool)
}

// ReadQC executes the Read Quorum Call on a configuration
// of Nodes and returns the most recent value.
func (c *Configuration) ReadQC(ctx context.Context, in *ReadRequest) (resp *ReadResponse, err error) {
	cd := gorums.QuorumCallData{
		Message: in,
		Method:  "storage.Storage.ReadQC",
	}
	cd.QuorumFunction = func(req protoreflect.ProtoMessage, replies map[uint32]protoreflect.ProtoMessage) (protoreflect.ProtoMessage, bool) {
		r := make(map[uint32]*ReadResponse, len(replies))
		for k, v := range replies {
			r[k] = v.(*ReadResponse)
		}
		return c.qspec.ReadQCQF(req.(*ReadRequest), r)
	}

	res, err := c.RawConfiguration.QuorumCall(ctx, cd)
	if err != nil {
		return nil, err
	}
	return res.(*ReadResponse), err
}

// WriteQC executes the Write Quorum Call on a configuration
// of Nodes and returns true if a majority of Nodes were updated.
func (c *Configuration) WriteQC(ctx context.Context, in *WriteRequest) (resp *WriteResponse, err error) {
	cd := gorums.QuorumCallData{
		Message: in,
		Method:  "storage.Storage.WriteQC",
	}
	cd.QuorumFunction = func(req protoreflect.ProtoMessage, replies map[uint32]protoreflect.ProtoMessage) (protoreflect.ProtoMessage, bool) {
		r := make(map[uint32]*WriteResponse, len(replies))
		for k, v := range replies {
			r[k] = v.(*WriteResponse)
		}
		return c.qspec.WriteQCQF(req.(*WriteRequest), r)
	}

	res, err := c.RawConfiguration.QuorumCall(ctx, cd)
	if err != nil {
		return nil, err
	}
	return res.(*WriteResponse), err
}

// ListKeysQC is a quorum call invoked on all nodes in configuration c,
// with the same argument in, and returns a combined result.
func (c *Configuration) ListKeysQC(ctx context.Context, in *ListRequest) (resp *ListResponse, err error) {
	cd := gorums.QuorumCallData{
		Message: in,
		Method:  "storage.Storage.ListKeysQC",
	}
	cd.QuorumFunction = func(req protoreflect.ProtoMessage, replies map[uint32]protoreflect.ProtoMessage) (protoreflect.ProtoMessage, bool) {
		r := make(map[uint32]*ListResponse, len(replies))
		for k, v := range replies {
			r[k] = v.(*ListResponse)
		}
		return c.qspec.ListKeysQCQF(req.(*ListRequest), r)
	}

	res, err := c.RawConfiguration.QuorumCall(ctx, cd)
	if err != nil {
		return nil, err
	}
	return res.(*ListResponse), err
}

// WriteMetaConfQC is a quorum call invoked on all nodes in configuration c,
// with the same argument in, and returns a combined result.
func (c *Configuration) WriteMetaConfQC(ctx context.Context, in *MetaConfig) (resp *WriteResponse, err error) {
	cd := gorums.QuorumCallData{
		Message: in,
		Method:  "storage.Storage.WriteMetaConfQC",
	}
	cd.QuorumFunction = func(req protoreflect.ProtoMessage, replies map[uint32]protoreflect.ProtoMessage) (protoreflect.ProtoMessage, bool) {
		r := make(map[uint32]*WriteResponse, len(replies))
		for k, v := range replies {
			r[k] = v.(*WriteResponse)
		}
		return c.qspec.WriteMetaConfQCQF(req.(*MetaConfig), r)
	}

	res, err := c.RawConfiguration.QuorumCall(ctx, cd)
	if err != nil {
		return nil, err
	}
	return res.(*WriteResponse), err
}

// ReadRPC executes the Read RPC on a single Node
func (n *Node) ReadRPC(ctx context.Context, in *ReadRequest) (resp *ReadResponse, err error) {
	cd := gorums.CallData{
		Message: in,
		Method:  "storage.Storage.ReadRPC",
	}

	res, err := n.RawNode.RPCCall(ctx, cd)
	if err != nil {
		return nil, err
	}
	return res.(*ReadResponse), err
}

// WriteRPC executes the Write RPC on a single Node
func (n *Node) WriteRPC(ctx context.Context, in *WriteRequest) (resp *WriteResponse, err error) {
	cd := gorums.CallData{
		Message: in,
		Method:  "storage.Storage.WriteRPC",
	}

	res, err := n.RawNode.RPCCall(ctx, cd)
	if err != nil {
		return nil, err
	}
	return res.(*WriteResponse), err
}

// ListKeysRPC is a quorum call invoked on all nodes in configuration c,
// with the same argument in, and returns a combined result.
func (n *Node) ListKeysRPC(ctx context.Context, in *ListRequest) (resp *ListResponse, err error) {
	cd := gorums.CallData{
		Message: in,
		Method:  "storage.Storage.ListKeysRPC",
	}

	res, err := n.RawNode.RPCCall(ctx, cd)
	if err != nil {
		return nil, err
	}
	return res.(*ListResponse), err
}

// Storage is the server-side API for the Storage Service
type Storage interface {
	ReadRPC(ctx gorums.ServerCtx, request *ReadRequest) (response *ReadResponse, err error)
	WriteRPC(ctx gorums.ServerCtx, request *WriteRequest) (response *WriteResponse, err error)
	ReadQC(ctx gorums.ServerCtx, request *ReadRequest) (response *ReadResponse, err error)
	WriteQC(ctx gorums.ServerCtx, request *WriteRequest) (response *WriteResponse, err error)
	WriteMulticast(ctx gorums.ServerCtx, request *WriteRequest)
	ListKeysRPC(ctx gorums.ServerCtx, request *ListRequest) (response *ListResponse, err error)
	ListKeysQC(ctx gorums.ServerCtx, request *ListRequest) (response *ListResponse, err error)
	WriteMetaConfQC(ctx gorums.ServerCtx, request *MetaConfig) (response *WriteResponse, err error)
}

func RegisterStorageServer(srv *gorums.Server, impl Storage) {
	srv.RegisterHandler("storage.Storage.ReadRPC", func(ctx gorums.ServerCtx, in *gorums.Message, finished chan<- *gorums.Message) {
		req := in.Message.(*ReadRequest)
		defer ctx.Release()
		resp, err := impl.ReadRPC(ctx, req)
		gorums.SendMessage(ctx, finished, gorums.WrapMessage(in.Metadata, resp, err))
	})
	srv.RegisterHandler("storage.Storage.WriteRPC", func(ctx gorums.ServerCtx, in *gorums.Message, finished chan<- *gorums.Message) {
		req := in.Message.(*WriteRequest)
		defer ctx.Release()
		resp, err := impl.WriteRPC(ctx, req)
		gorums.SendMessage(ctx, finished, gorums.WrapMessage(in.Metadata, resp, err))
	})
	srv.RegisterHandler("storage.Storage.ReadQC", func(ctx gorums.ServerCtx, in *gorums.Message, finished chan<- *gorums.Message) {
		req := in.Message.(*ReadRequest)
		defer ctx.Release()
		resp, err := impl.ReadQC(ctx, req)
		gorums.SendMessage(ctx, finished, gorums.WrapMessage(in.Metadata, resp, err))
	})
	srv.RegisterHandler("storage.Storage.WriteQC", func(ctx gorums.ServerCtx, in *gorums.Message, finished chan<- *gorums.Message) {
		req := in.Message.(*WriteRequest)
		defer ctx.Release()
		resp, err := impl.WriteQC(ctx, req)
		gorums.SendMessage(ctx, finished, gorums.WrapMessage(in.Metadata, resp, err))
	})
	srv.RegisterHandler("storage.Storage.WriteMulticast", func(ctx gorums.ServerCtx, in *gorums.Message, _ chan<- *gorums.Message) {
		req := in.Message.(*WriteRequest)
		defer ctx.Release()
		impl.WriteMulticast(ctx, req)
	})
	srv.RegisterHandler("storage.Storage.ListKeysRPC", func(ctx gorums.ServerCtx, in *gorums.Message, finished chan<- *gorums.Message) {
		req := in.Message.(*ListRequest)
		defer ctx.Release()
		resp, err := impl.ListKeysRPC(ctx, req)
		gorums.SendMessage(ctx, finished, gorums.WrapMessage(in.Metadata, resp, err))
	})
	srv.RegisterHandler("storage.Storage.ListKeysQC", func(ctx gorums.ServerCtx, in *gorums.Message, finished chan<- *gorums.Message) {
		req := in.Message.(*ListRequest)
		defer ctx.Release()
		resp, err := impl.ListKeysQC(ctx, req)
		gorums.SendMessage(ctx, finished, gorums.WrapMessage(in.Metadata, resp, err))
	})
	srv.RegisterHandler("storage.Storage.WriteMetaConfQC", func(ctx gorums.ServerCtx, in *gorums.Message, finished chan<- *gorums.Message) {
		req := in.Message.(*MetaConfig)
		defer ctx.Release()
		resp, err := impl.WriteMetaConfQC(ctx, req)
		gorums.SendMessage(ctx, finished, gorums.WrapMessage(in.Metadata, resp, err))
	})
}

type internalListResponse struct {
	nid   uint32
	reply *ListResponse
	err   error
}

type internalReadResponse struct {
	nid   uint32
	reply *ReadResponse
	err   error
}

type internalWriteResponse struct {
	nid   uint32
	reply *WriteResponse
	err   error
}
