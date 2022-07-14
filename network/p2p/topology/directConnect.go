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

import "github.com/SealSC/SealABC/network/p2p"

//type TopologyRegister func(IProcessor) (error)

type DirectConnect struct {
	LocalNode p2p.LinkNode
}

func (t *DirectConnect) Name() string {
	return "direct connect"
}

func (t *DirectConnect) MountTo(router p2p.IRouter) {
	return
}

func (t *DirectConnect) BuildNodeID(node p2p.Node) string {
	return node.ServeAddress
}

func (t *DirectConnect) InterestedMessage(msg p2p.Message) (interested bool) {
	return false
}

func (t *DirectConnect) MessageProcessor(msg p2p.Message, link p2p.ILink) {
	return
}

func (t *DirectConnect) Join(node p2p.LinkNode) (err error) {
	return
}

func (t *DirectConnect) Leave() {
	return
}

func (t *DirectConnect) SetLocalNode(node p2p.LinkNode) {
	t.LocalNode = node
}

func (t *DirectConnect) GetLocalNode() (node p2p.LinkNode) {
	return t.LocalNode
}

func (t *DirectConnect) GetLink(node p2p.Node) (link p2p.ILink, err error) {
	return
}

func (t *DirectConnect) GetAllNodes() (all []p2p.LinkNode) {
	return
}

func (t *DirectConnect) AddLink(link p2p.ILink) {

}

func (t *DirectConnect) RemoveLink(link p2p.ILink) {

}
