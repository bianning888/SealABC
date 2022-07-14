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

package topology

import (
	"encoding/json"
	"errors"
	"github.com/SealSC/SealABC/crypto/signers/ed25519"
	"github.com/SealSC/SealABC/log"
	"github.com/SealSC/SealABC/network/p2p"
	"github.com/SealSC/SealABC/network/p2p/topology/p2p/fullyConnect/message"
	"github.com/SealSC/SealABC/network/p2p/topology/p2p/fullyConnect/message/payload"
)

type Topology struct {
	LocalNode           p2p.LinkNode
	preJoinNode         map[p2p.ILink]p2p.LinkNode
	joinedNode          map[p2p.ILink]p2p.LinkNode
	nodeID2Link         map[string]p2p.ILink
	messageProcessorMap map[string]iMessageProcessor
	router              p2p.IRouter
}

func (t Topology) Name() string {
	return "fully connected P2P"
}

func (t *Topology) MountTo(router p2p.IRouter) {
	message.LoadMessageTypes()

	t.router = router

	t.preJoinNode = map[p2p.ILink]p2p.LinkNode{}
	t.joinedNode = map[p2p.ILink]p2p.LinkNode{}
	t.nodeID2Link = map[string]p2p.ILink{}

	t.messageProcessorMap = map[string]iMessageProcessor{
		message.Types.Join.String():              JoinMessageProcessor,
		message.Types.JoinReply.String():         JoinReplyMessageProcessor,
		message.Types.GetNeighbors.String():      GetNeighborsMessageProcessor,
		message.Types.GetNeighborsReply.String(): GetNeighborsReplyMessageProcessor,
		message.Types.Ping.String():              PingMessageProcessor,
		message.Types.Pong.String():              PongMessageProcessor,
	}
}

func (t *Topology) BuildNodeID(_ p2p.Node) string {
	//todo: create random id, not key
	s, _ := ed25519.SignerGenerator.NewSigner(nil)
	return s.PublicKeyString()
}

func (t *Topology) InterestedMessage(msg p2p.Message) (interested bool) {
	return msg.Family == message.Family
}

func (t *Topology) MessageProcessor(msg p2p.Message, link p2p.ILink) {
	processor, exists := t.messageProcessorMap[msg.Type]
	if !exists {
		return
	}

	err := processor.Process(msg, t, link)

	if err != nil {
		log.Log.Println("got p2p error: ", err)
	}

	return
}

func (t *Topology) SetLocalNode(node p2p.LinkNode) {
	t.LocalNode = node
}

func (t *Topology) GetLocalNode() (node p2p.LinkNode) {
	return t.LocalNode
}

func (t *Topology) Join(node p2p.LinkNode) (err error) {
	t.preJoinNode[node.Link] = node
	join := payload.Join{
		TargetID: node.ID,
		SourceID: t.LocalNode.ID,
	}

	joinPayload, _ := json.Marshal(join)
	joinMsg := message.NewMessage(message.Types.Join, joinPayload)

	joinMsg.From = t.LocalNode.Node
	_, err = node.Link.SendMessage(joinMsg)
	if err != nil {
		log.Log.Warn("join to seed failed: ", node.ServeAddress)
	} else {
		log.Log.Println("send join message to seed success: ", node.ServeAddress)
	}
	return
}

func (t *Topology) Leave() {
	return
}

func (t *Topology) getLinkFromPreJoinList(node p2p.Node) (link p2p.ILink) {
	for l, n := range t.preJoinNode {
		if n.ServeAddress == node.ServeAddress && n.Protocol == node.Protocol {
			link = l
			break
		}
	}

	return
}

func (t *Topology) GetLink(node p2p.Node) (link p2p.ILink, err error) {
	link, exists := t.nodeID2Link[node.ID]
	if exists {
		return
	}

	link = t.getLinkFromPreJoinList(node)
	if link != nil {
		return
	}
	err = errors.New("no such link")
	return
}

func (t *Topology) GetAllNodes() (all []p2p.LinkNode) {
	for _, n := range t.joinedNode {
		if n.ID == t.LocalNode.ID {
			continue
		}

		all = append(all, n)
	}

	return
}

func (t *Topology) AddLink(link p2p.ILink) {
	node := p2p.NewNetworkNodeFromLink(link)
	t.preJoinNode[link] = node
}

func (t *Topology) RemoveLink(link p2p.ILink) {
	node, exist := t.getNode(link)
	if exist {
		delete(t.joinedNode, link)
		delete(t.nodeID2Link, node.ID)
	}
	delete(t.preJoinNode, link)
}

func (t *Topology) getPreJoinNode(link p2p.ILink) (node p2p.LinkNode, exist bool) {
	node, exist = t.preJoinNode[link]
	return
}

func (t *Topology) getNode(link p2p.ILink) (linkedNode p2p.LinkNode, exist bool) {
	linkedNode, exist = t.joinedNode[link]
	return
}

func (t *Topology) setJoinedNode(node p2p.LinkNode) {
	if node.ID == t.LocalNode.ID {
		return
	}

	t.joinedNode[node.Link] = node
	t.nodeID2Link[node.ID] = node.Link
	delete(t.preJoinNode, node.Link)
}

func (t *Topology) removeNode(node p2p.LinkNode) {
}

func (t *Topology) isJoined(id string) (joined bool) {
	if id == t.LocalNode.ID {
		return true
	}
	_, joined = t.nodeID2Link[id]
	return
}
