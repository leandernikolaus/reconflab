package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"reconfstorage/proto"

	"github.com/relab/gorums"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type client struct {
	mgr  *proto.Manager
	cfg  *proto.Configuration
	pcfg *proto.MetaConfig
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

	log.Println("Manager created")
	log.Printf("Adddresses %s\n", addresses)
	// create configuration containing all nodes
	cfg, err := mgr.NewConfiguration(&qspec{cfgSize: len(addresses)}, gorums.WithNodeList(addresses))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Conf created")

	pcfg := &proto.MetaConfig{Adds: "0:" + fmt.Sprint(len(addresses)-1), Time: &timestamppb.Timestamp{Seconds: 1, Nanos: 1}}

	return &client{
		mgr:  mgr,
		cfg:  cfg,
		pcfg: pcfg,
	}
}

// find config with minimal timestamp
// return its key
func getMin(configs map[string]*proto.MetaConfig) string {
	if len(configs) == 0 {
		return ""
	}

	min := ""

	for s, config := range configs {
		if min == "" || config.Time.AsTime().After(configs[min].Time.AsTime()) {
			min = s
		}
	}
	return min
}

// addConfigs checks newconfigs for once that are newer than cur
// newer confgs are added to the confmap
// if a newer started configuration is found, this is the only returned function
// and the client state is updated
func (c *client) addConfigs(confmap map[string]*proto.MetaConfig, cur *proto.MetaConfig, newconfigs []*proto.MetaConfig) map[string]*proto.MetaConfig {
	for _, cc := range newconfigs {
		if TimeBefore(cur.GetTime(), cc.GetTime()) {
			if cc.GetStarted() {
				//empty map
				confmap = make(map[string]*proto.MetaConfig, 1)

				//update client state
				c.pcfg = cc

			}
			confmap[cc.GetTime().String()] = cc
		}
	}
	return confmap
}

func TimeBefore(a, b *timestamppb.Timestamp) bool {
	return a.AsTime().Before(b.AsTime())
}

func (c *client) read(key string) *proto.ReadResponse {
	confmap := map[string]*proto.MetaConfig{c.pcfg.Time.String(): c.pcfg}
	resp := &proto.ReadResponse{Time: &timestamppb.Timestamp{Seconds: 0, Nanos: 0}}

	for min := getMin(confmap); len(confmap) > 0; {
		cfg, err := c.parseConfiguration(confmap[min].Adds)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		minresp := c.readQC(key, cfg)

		// remember Value, if it has larger Time
		if TimeBefore(resp.GetTime(), minresp.GetTime()) {
			resp = minresp
		}

		confmap = c.addConfigs(confmap, confmap[min], minresp.GetMConfigs())
		delete(confmap, min)
	}
	return resp
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

	// TODO perform the write on c.cfg and all its successors

	return &proto.WriteResponse{New: false}
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
	confmap := map[string]*proto.MetaConfig{c.pcfg.Time.String(): c.pcfg}

	var keys map[string]bool

	for min := getMin(confmap); len(confmap) > 0; {
		cfg, err := c.parseConfiguration(confmap[min].Adds)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		minresp := c.listQC(cfg)

		if len(keys) == 0 {
			keys = make(map[string]bool, len(minresp.GetKeys()))
		}
		for _, k := range minresp.GetKeys() {
			keys[k] = true
		}

		confmap = c.addConfigs(confmap, confmap[min], minresp.GetMConfigs())
		delete(confmap, min)

	}

	allkeys := make([]string, 0, len(keys))
	for k := range keys {
		allkeys = append(allkeys, k)
	}

	return &proto.ListResponse{Keys: allkeys}
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

func (c *client) writeConfig(target *proto.MetaConfig) *proto.WriteResponse {
	confmap := map[string]*proto.MetaConfig{c.pcfg.Time.String(): c.pcfg}

	for min := getMin(confmap); len(confmap) > 0; {

		if target.GetTime().AsTime().Before(confmap[min].GetTime().AsTime()) {
			return &proto.WriteResponse{New: false}
		}

		cfg, err := c.parseConfiguration(confmap[min].Adds)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		minresp := c.writeConfigQC(target, cfg)

		confmap = c.addConfigs(confmap, confmap[min], minresp.GetMConfigs())
		delete(confmap, min)

	}

	return &proto.WriteResponse{New: true}
}

func (client) writeConfigQC(conf *proto.MetaConfig, cfg *proto.Configuration) *proto.WriteResponse {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := cfg.WriteMetaConfQC(ctx, conf)
	cancel()
	if err != nil {
		fmt.Printf("ListKeys RPC finished with error: %v\n", err)
		return nil
	}
	return resp
}

func (c *client) reconf(newAdds string) {

	goalProtoConf := &proto.MetaConfig{Adds: newAdds, Started: false, Time: timestamppb.Now()}
	// create a Configuration used for quorum calls.
	goalCfg, err := c.parseConfiguration(newAdds)
	if err != nil {
		log.Printf("Could not create new configuration: %s", newAdds)
		return
	}

	// TODO: perform reconfiguration

	goalProtoConf.Started = true

	// update the default configuration used by the client
	c.cfg = goalCfg
	c.pcfg = goalProtoConf
}

func (c client) parseConfiguration(cfgStr string) (cfg *proto.Configuration, err error) {
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
				return nil, err
			}
		}
		if i == len(cfgStr)-1 {
			stop = numNodes
		} else {
			stop, err = strconv.Atoi(cfgStr[i+1:])
			if err != nil {
				fmt.Printf("Failed to parse configuration: %v\n", err)
				return nil, err
			}
		}
		if start >= stop || start < 0 || stop > numNodes {
			fmt.Println("Invalid configuration.")
			return nil, fmt.Errorf("could not parse configuration: %s", cfgStr)
		}
		nodes := make([]string, 0)
		for _, node := range c.mgr.Nodes()[start:stop] {
			nodes = append(nodes, node.Address())
		}
		cfg, err = c.mgr.NewConfiguration(&qspec{cfgSize: stop - start}, gorums.WithNodeList(nodes))
		if err != nil {
			fmt.Printf("Failed to create configuration: %v\n", err)
			return nil, err
		}
		return cfg, nil
	}
	// configuration using list of indices
	if indices := strings.Split(cfgStr, ","); len(indices) > 0 {
		selectedNodes := make([]string, 0, len(indices))
		nodes := c.mgr.Nodes()
		for _, index := range indices {
			i, err := strconv.Atoi(index)
			if err != nil {
				fmt.Printf("Failed to parse configuration: %v\n", err)
				return nil, err
			}
			if i < 0 || i >= len(nodes) {
				fmt.Println("Invalid configuration.")
				return nil, fmt.Errorf("could not parse configuration: %s", cfgStr)
			}
			selectedNodes = append(selectedNodes, nodes[i].Address())
		}
		cfg, err := c.mgr.NewConfiguration(&qspec{cfgSize: len(selectedNodes)}, gorums.WithNodeList(selectedNodes))
		if err != nil {
			fmt.Printf("Failed to create configuration: %v\n", err)
			return nil, err
		}
		return cfg, nil
	}
	fmt.Println("Invalid configuration.")
	return nil, fmt.Errorf("could not parse configuration: %s", cfgStr)
}
