package comm

type Node struct {
	NodeID uint32
	Addr   string
}

func GetNodeByID(nodes []Node, nodeID uint32) *Node {
	var node *Node
	for i := range nodes {
		if nodes[i].NodeID == nodeID {
			node = &nodes[i]
		}
	}
	return node
}

type Nodes interface {
	GetMyNode() *Node
	GetNodes() []Node
	GetMajorityCount() uint
}
