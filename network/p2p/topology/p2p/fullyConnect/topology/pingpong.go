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
	"github.com/SealSC/SealABC/network/p2p"
	"github.com/SealSC/SealABC/network/p2p/topology/p2p/fullyConnect/message"
	payload2 "github.com/SealSC/SealABC/network/p2p/topology/p2p/fullyConnect/message/payload"
	"time"
)

func doPing(link p2p.ILink) {
	ping := payload2.NewPing()
	pingPayloadBytes, _ := json.Marshal(ping)
	msg := message.NewMessage(message.Types.Ping, pingPayloadBytes)
	link.SendMessage(msg)
}

type pingMessageProcessor struct{}

func (p *pingMessageProcessor) Process(msg p2p.Message, t *Topology, link p2p.ILink) (err error) {
	pingPayload := payload2.PingPongPayload{}
	err = payload2.FromMessage(msg, &pingPayload)
	if err != nil {
		return
	}

	pong := payload2.PingPongPayload{
		Number: pingPayload.Number,
	}

	replayPayloadBytes, _ := json.Marshal(pong)
	reply := message.NewMessage(message.Types.Pong, replayPayloadBytes)
	link.SendMessage(reply)

	go func() {
		time.Sleep(time.Second * 10)

		doPing(link)
	}()

	return
}

type pongMessageProcessor struct{}

func (p *pongMessageProcessor) Process(msg p2p.Message, topology *Topology, _ p2p.ILink) (err error) {
	//todo: refresh neighbor's state

	return
}

var PingMessageProcessor = &pingMessageProcessor{}
var PongMessageProcessor = &pongMessageProcessor{}
