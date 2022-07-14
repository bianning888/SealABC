package p2pv2

import (
	"github.com/SealSC/SealABC/metadata/message"
	"github.com/SealSC/SealABC/network/p2p"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
)

const (
	TopologyName = ""
)

type P2PServer struct {
	host host.Host
}

func (p *P2PServer) Self() p2p.Node {
	//TODO implement me
	panic("implement me")
}

func (p *P2PServer) TopologyName() string {
	return TopologyName
}

func (p *P2PServer) Start(cfg p2p.Config) (err error) {
	opts, err := p.createLibp2pOptions(cfg)

	h, err := libp2p.New(opts...)
	if err != nil {
		return err
	}

	p.host = h
	return
}

func (p *P2PServer) ConnectTo(node p2p.Node) (linkedNode p2p.LinkNode, err error) {
	//TODO implement me
	panic("implement me")
}

func (p *P2PServer) LinkClosed(link p2p.ILink) {
	//TODO implement me
	panic("implement me")
}

func (p *P2PServer) JoinTopology(seed p2p.Node) (err error) {
	//TODO implement me
	panic("implement me")
}

func (p *P2PServer) LeaveTopology() {
	//TODO implement me
	panic("implement me")
}

func (p *P2PServer) GetAllLinkedNode() (nodes []p2p.Node) {
	//TODO implement me
	panic("implement me")
}

func (p *P2PServer) RawMessageProcessor(data []byte, link p2p.ILink) {
	//TODO implement me
	panic("implement me")
}

func (p *P2PServer) RegisterMessageProcessor(msgFamily string, processor p2p.MessageProcessor) {
	//TODO implement me
	panic("implement me")
}

func (p *P2PServer) SendTo(node p2p.Node, msg p2p.Message) (n int, err error) {
	//TODO implement me
	panic("implement me")
}

func (p *P2PServer) Broadcast(msg message.Message) (err error) {
	//TODO implement me
	panic("implement me")
}

func (p *P2PServer) createLibp2pOptions(cfg p2p.Config) ([]libp2p.Option, error) {

	var opts []libp2p.Option

	pk, err := getPrivateKey(cfg.SelfSigner)
	if err != nil {
		return opts, err
	}

	opts = append(opts, libp2p.Identity(pk))

	return opts, nil
}
