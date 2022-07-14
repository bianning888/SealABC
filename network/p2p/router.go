package p2p

import (
	"github.com/SealSC/SealABC/metadata/message"
)

type RawMessageProcessor func(data []byte, link ILink)
type MessageProcessor func(msg Message) (reply *Message)
type LinkClosed func(link ILink)

type IRouter interface {
	Self() Node

	TopologyName() string

	Start(cfg Config) (err error)
	//Listen(listener net.Listener)
	ConnectTo(node Node) (linkedNode LinkNode, err error)

	LinkClosed(link ILink)

	JoinTopology(seed Node) (err error)
	LeaveTopology()
	GetAllLinkedNode() (nodes []Node)

	RawMessageProcessor(data []byte, link ILink)
	RegisterMessageProcessor(msgFamily string, processor MessageProcessor)

	SendTo(node Node, msg Message) (n int, err error)
	Broadcast(msg message.Message) (err error)
}
