/*
 * Copyright 2020 The SealABC Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package p2p

import (
	"github.com/SealSC/SealABC/log"
	"github.com/SealSC/SealABC/metadata/message"
	"github.com/SealSC/SealABC/network/p2p"
	"github.com/SealSC/SealABC/network/p2p/topology"
	"net"
	"sync"
)

type Router struct {
	Topology            p2p.ITopology
	MessageProcessorMap map[string]p2p.MessageProcessor
	LocalNode           p2p.LinkNode

	rawProcessorLock sync.Mutex
}

func (r *Router) Self() p2p.Node {
	return r.LocalNode.Node
}

func (r *Router) TopologyName() string {
	return r.Topology.Name()
}

func (r *Router) Start(cfg p2p.Config) (err error) {
	r.MessageProcessorMap = map[string]p2p.MessageProcessor{}
	if cfg.Topology != nil {
		r.Topology = cfg.Topology
	} else {
		r.Topology = &topology.DirectConnect{}
	}

	r.Topology.MountTo(r)

	localNode := p2p.LinkNode{}
	localNode.Protocol = cfg.ServiceProtocol
	localNode.ServeAddress = cfg.ServiceAddress
	if cfg.ID == "" {
		localNode.ID = r.Topology.BuildNodeID(localNode.Node)
	} else {
		localNode.ID = cfg.ID
	}

	r.LocalNode = localNode
	r.Topology.SetLocalNode(localNode)

	log.Log.Println("[ I am ]: ", localNode.ID)

	var listener net.Listener
	if !cfg.ClientOnly {
		listener, err = net.Listen(cfg.ServiceProtocol, cfg.ServiceAddress)
		if err != nil {
			return
		}

		go r.Listen(listener)
	}

	return
}

func (r *Router) Listen(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Log.Println("accept error:", err)
			break
		}

		newLink := Link{
			Connection:          conn,
			ConnectOut:          false,
			RawMessageProcessor: r.RawMessageProcessor,
			LinkClosed:          r.LinkClosed,
		}

		r.Topology.AddLink(&newLink)
		newLink.Start()
	}
}

func (r *Router) ConnectTo(node p2p.Node) (linkedNode p2p.LinkNode, err error) {
	conn, err := net.Dial(node.Protocol, node.ServeAddress)
	if err != nil {
		log.Log.Println("got an error: ", err)
		return
	}

	link := Link{
		Connection:          conn,
		ConnectOut:          true,
		RawMessageProcessor: r.RawMessageProcessor,
		LinkClosed:          r.LinkClosed,
	}
	link.Start()

	linkedNode = p2p.NewNetworkNodeFromLink(&link)
	linkedNode.ServeAddress = node.ServeAddress
	linkedNode.ID = r.Topology.BuildNodeID(linkedNode.Node)

	return
}

func (r *Router) LinkClosed(link p2p.ILink) {
	r.Topology.RemoveLink(link)
	return
}

func (r *Router) JoinTopology(seed p2p.Node) (err error) {
	if seed.ID == "" {
		seed.ID = r.Topology.BuildNodeID(seed)
	}

	if _, err = r.Topology.GetLink(seed); err == nil {
		return
	}

	linkedNode, err := r.ConnectTo(seed)
	if err != nil {
		return
	}

	log.Log.Println("connect to seed: ", seed)
	err = r.Topology.Join(linkedNode)

	return
}

func (r *Router) LeaveTopology() {
	r.Topology.Leave()
}

func (r *Router) GetAllLinkedNode() (nodes []p2p.Node) {
	linkedNodes := r.Topology.GetAllNodes()
	for _, n := range linkedNodes {
		nodes = append(nodes, n.Node)
	}

	return
}

func (r *Router) RegisterMessageProcessor(msgFamily string, processor p2p.MessageProcessor) {
	r.MessageProcessorMap[msgFamily] = processor
}

func (r *Router) RawMessageProcessor(data []byte, link p2p.ILink) {
	r.rawProcessorLock.Lock()
	defer r.rawProcessorLock.Unlock()

	newMsg := p2p.Message{}
	err := newMsg.FromRawMessage(data)
	if err != nil {
		return
	}

	if r.Topology.InterestedMessage(newMsg) {
		r.Topology.MessageProcessor(newMsg, link)
	}

	msgProcessor, exists := r.MessageProcessorMap[newMsg.Family]
	if !exists {
		return
	}

	replyMsg := msgProcessor(newMsg)
	if replyMsg == nil {
		return
	}

	replyMsg.From = r.LocalNode.Node
	rawMsg, err := replyMsg.ToRawMessage()
	if err != nil {
		return
	}

	n, err := link.SendData(rawMsg)

	if err != nil {
		log.Log.Println("reply message failed. ", err)
		log.Log.Printf("\r\nshould sent %d real sent %d \r\n", len(rawMsg), n)
	}

	rawMsg = nil
}

func (r *Router) SendTo(node p2p.Node, msg p2p.Message) (n int, err error) {
	link, err := r.Topology.GetLink(node)
	if err != nil {
		return
	}

	msg.From = r.LocalNode.Node
	rawMsg, err := msg.ToRawMessage()
	if err != nil {
		return
	}

	n, err = link.SendData(rawMsg)
	if err != nil {
		log.Log.Error("send data failed: data length ", len(rawMsg), " error: ", err.Error())
	}

	return
}

func (r *Router) Broadcast(msg message.Message) (err error) {
	networkMsg := p2p.Message{
		Message: msg,
		From:    r.LocalNode.Node,
	}

	targets := r.Topology.GetAllNodes()
	for _, t := range targets {
		n := t.Node
		go r.SendTo(n, networkMsg)
	}
	return
}
