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
	"sync"
)

func getNeighbors(seed p2p.LinkNode) {
	log.Log.Println("get neighbors from: ", seed.ID)
	msg := message.NewMessage(message.Types.GetNeighbors, []byte{})
	_, err := seed.Link.SendMessage(msg)
	if err != nil {
		log.Log.Error("send get neighbors message failed: ", err.Error())
	}
}

type getNeighborsMessageProcessor struct{}

func (j *getNeighborsMessageProcessor) Process(msg p2p.Message, t *Topology, link p2p.ILink) (err error) {
	allNodes := t.GetAllNodes()
	neighbors := payload2.NeighborsPayload{}

	for _, n := range allNodes {
		neighbors.Neighbors = append(neighbors.Neighbors, n.Node)
	}

	replyPayload, _ := json.Marshal(neighbors)
	replyMsg := message.NewMessage(message.Types.GetNeighborsReply, replyPayload)

	_, err = link.SendMessage(replyMsg)

	return
}

var GetNeighborsMessageProcessor = &getNeighborsMessageProcessor{}

type getNeighborsReplyMessageProcessor struct {
	joinLock sync.Mutex
}

func (j *getNeighborsReplyMessageProcessor) Process(msg p2p.Message, topology *Topology, link p2p.ILink) (err error) {
	neighbors := payload2.NeighborsPayload{}
	err = payload2.FromMessage(msg, &neighbors)
	if err != nil {
		log.Log.Warn("invalid get neighbor reply message: ", err.Error())
		return err
	}

	for _, n := range neighbors.Neighbors {
		if topology.isJoined(n.ID) {
			continue
		}

		j.joinLock.Lock()
		err := topology.router.JoinTopology(n)
		j.joinLock.Unlock()

		if err != nil {
			log.Log.Warn("join neighbor: ", n, " failed.")
			continue
		}
	}

	return
}

var GetNeighborsReplyMessageProcessor = &getNeighborsReplyMessageProcessor{}
