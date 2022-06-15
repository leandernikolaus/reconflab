package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"storage/proto"

	"github.com/relab/gorums"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type client struct {
	mgr *proto.Manager
	cfg *proto.Configuration
}

func newClient(addresses []string) *client {
	if len(addresses) < 1 {
		log.Fatalln("No addresses provided!")
	}

	// init gorums manager
	mgr := proto.NewManager(
		gorums.WithDialTimeout(1*time.Second),
		gorums.WithGrpcDialOptions(
			grpc.WithBlock(), // block until connections are made
			grpc.WithTransportCredentials(insecure.NewCredentials()), // disable TLS
		),
	)
	// create configuration containing all nodes
	cfg, err := mgr.NewConfiguration(&qspec{cfgSize: len(addresses)}, gorums.WithNodeList(addresses))
	if err != nil {
		log.Fatal(err)
	}

	return &client{
		mgr: mgr,
		cfg: cfg,
	}
}

func (c *client) read(key string) *proto.ReadResponse {
	return c.readQC(key, c.cfg)
}

func (client) readQC(key string, cfg *proto.Configuration) *proto.ReadResponse {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := cfg.ReadQC(ctx, &proto.ReadRequest{Key: key})
	cancel()
	if err != nil {
		fmt.Printf("Read RPC finished with error: %v\n", err)
		return nil
	}
	return resp
}

func (c *client) write(key, value string) *proto.WriteResponse {
	return c.writeQC(key, value, c.cfg)
}

func (client) writeQC(key, value string, cfg *proto.Configuration) *proto.WriteResponse {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := cfg.WriteQC(ctx, &proto.WriteRequest{Key: key, Value: value, Time: timestamppb.Now()})
	cancel()
	if err != nil {
		fmt.Printf("Write RPC finished with error: %v\n", err)
		return nil
	}
	return resp
}

func (c *client) list() *proto.ListResponse {
	return c.listQC(c.cfg)
}

func (client) listQC(cfg *proto.Configuration) *proto.ListResponse {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := cfg.ListKeysQC(ctx, &proto.ListRequest{})
	cancel()
	if err != nil {
		fmt.Printf("ListKeys RPC finished with error: %v\n", err)
		return nil
	}
	return resp
}

func (c client) parseConfiguration(cfgStr string) (cfg *proto.Configuration) {
	// configuration using range syntax
	if i := strings.Index(cfgStr, ":"); i > -1 {
		var start, stop int
		var err error
		numNodes := c.mgr.Size()
		if i == 0 {
			start = 0
		} else {
			start, err = strconv.Atoi(cfgStr[:i])
			if err != nil {
				fmt.Printf("Failed to parse configuration: %v\n", err)
				return nil
			}
		}
		if i == len(cfgStr)-1 {
			stop = numNodes
		} else {
			stop, err = strconv.Atoi(cfgStr[i+1:])
			if err != nil {
				fmt.Printf("Failed to parse configuration: %v\n", err)
				return nil
			}
		}
		if start >= stop || start < 0 || stop >= numNodes {
			fmt.Println("Invalid configuration.")
			return nil
		}
		nodes := make([]string, 0)
		for _, node := range c.mgr.Nodes()[start:stop] {
			nodes = append(nodes, node.Address())
		}
		cfg, err = c.mgr.NewConfiguration(&qspec{cfgSize: stop - start}, gorums.WithNodeList(nodes))
		if err != nil {
			fmt.Printf("Failed to create configuration: %v\n", err)
			return nil
		}
		return cfg
	}
	// configuration using list of indices
	if indices := strings.Split(cfgStr, ","); len(indices) > 0 {
		selectedNodes := make([]string, 0, len(indices))
		nodes := c.mgr.Nodes()
		for _, index := range indices {
			i, err := strconv.Atoi(index)
			if err != nil {
				fmt.Printf("Failed to parse configuration: %v\n", err)
				return nil
			}
			if i < 0 || i >= len(nodes) {
				fmt.Println("Invalid configuration.")
				return nil
			}
			selectedNodes = append(selectedNodes, nodes[i].Address())
		}
		cfg, err := c.mgr.NewConfiguration(&qspec{cfgSize: len(selectedNodes)}, gorums.WithNodeList(selectedNodes))
		if err != nil {
			fmt.Printf("Failed to create configuration: %v\n", err)
			return nil
		}
		return cfg
	}
	fmt.Println("Invalid configuration.")
	return nil
}
