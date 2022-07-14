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
	"github.com/SealSC/SealABC/log"
	"github.com/SealSC/SealABC/network/p2p"
	"github.com/SealSC/SealABC/network/p2p/topology/p2p/fullyConnect/message"
	payload2 "github.com/SealSC/SealABC/network/p2p/topology/p2p/fullyConnect/message/payload"
)

type joinMessageProcessor struct{}

func (j *joinMessageProcessor) Process(msg p2p.Message, t *Topology, link p2p.ILink) (err error) {
	join := payload2.Join{}
	err = payload2.FromMessage(msg, &join)
	if err != nil {
		log.Log.Println("not join protocol ")
		return
	}

	joinReply := payload2.JoinReply{
		PrevID: join.TargetID,
		RealID: t.LocalNode.ID,
	}
	replyPayload, _ := json.Marshal(joinReply)

	target, exist := t.getPreJoinNode(link)

	if !exist {
		return
	}

	log.Log.Println("got join message from: ", msg.From)
	target.Node = msg.From
	t.setJoinedNode(target)

	reply := message.NewMessage(message.Types.JoinReply, replyPayload)
	reply.From = t.LocalNode.Node
	rawReply, _ := reply.ToRawMessage()
	_, err = target.Link.SendData(rawReply)
	if err != nil {
		log.Log.Warn("send join reply failed: ", err.Error())
	}
	return
}

type joinReplyMessageProcessor struct{}

func (j *joinReplyMessageProcessor) Process(msg p2p.Message, t *Topology, link p2p.ILink) (err error) {
	joinReply := payload2.JoinReply{}
	err = payload2.FromMessage(msg, &joinReply)

	if err != nil {
		log.Log.Println("not join-reply protocol ")
		return
	}

	target, exist := t.getPreJoinNode(link)

	if !exist {
		return
	}

	log.Log.Println("got joinReply message from: ", msg.From)
	target.Node = msg.From
	t.setJoinedNode(target)

	//doPing(link)

	//get neighbors
	getNeighbors(target)
	return
}

var JoinMessageProcessor = &joinMessageProcessor{}
var JoinReplyMessageProcessor = &joinReplyMessageProcessor{}
