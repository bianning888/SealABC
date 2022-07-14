package p2p

type ITopology interface {
	Name() string

	MountTo(router IRouter)
	BuildNodeID(node Node) string

	InterestedMessage(msg Message) (interested bool)
	MessageProcessor(msg Message, link ILink)

	Join(seed LinkNode) (err error)
	Leave()

	SetLocalNode(node LinkNode)
	GetLink(node Node) (link ILink, err error)
	GetAllNodes() []LinkNode

	AddLink(link ILink)
	RemoveLink(link ILink)
}
