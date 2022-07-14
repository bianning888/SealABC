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

package chainNetwork

import (
	"errors"
	"github.com/SealSC/SealABC/dataStructure/enum"
	"github.com/SealSC/SealABC/log"
	"github.com/SealSC/SealABC/network/p2p"
	"github.com/SealSC/SealABC/service/system/blockchain/chainStructure"
	"sync"
)

type P2PService struct {
	syncLock              sync.Mutex
	chain                 *chainStructure.Blockchain
	networkMessageHandler map[string]p2pMessageHandler

	//export
	NetworkService p2p.IService
}

func startChainP2PNetwork(cfg p2p.Config, ps *P2PService) (networkService p2p.IService, err error) {
	networkService = &p2p.Service{}
	err = networkService.Create(cfg)
	if err != nil {
		return
	}

	if len(cfg.P2PSeeds) == 0 {
		err = errors.New("no p2p seeds")
		return
	}

	var seedNodes []p2p.Node
	for _, seed := range cfg.P2PSeeds {
		newP2PNode := p2p.Node{}

		newP2PNode.ServeAddress = seed
		newP2PNode.Protocol = cfg.ServiceProtocol

		seedNodes = append(seedNodes, newP2PNode)
	}

	err = networkService.Join(seedNodes, nil)

	networkService.RegisterMessageProcessor(messageFamily, ps.handleP2PMessage)
	return
}

func Load() {
	enum.SimpleBuild(&MessageTypes)
}

func NewNetwork(cfg p2p.Config, chain *chainStructure.Blockchain) *P2PService {
	p2p := P2PService{}
	p2p.networkMessageHandler = map[string]p2pMessageHandler{
		MessageTypes.PushRequest.String():    p2p.handlePushRequest,
		MessageTypes.SyncBlock.String():      p2p.handleSyncBlock,
		MessageTypes.SyncBlockReply.String(): p2p.handleSyncBlockReply,
	}

	ns, err := startChainP2PNetwork(cfg, &p2p)
	if err != nil {
		log.Log.Warn("blockchain service network started with an error: ", err.Error())
	}

	p2p.chain = chain
	p2p.NetworkService = ns

	return &p2p
}
