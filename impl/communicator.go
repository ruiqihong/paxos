package impl

import (
	"github.com/ruiqihong/paxos/coder"
	"github.com/ruiqihong/paxos/comm"
	"github.com/ruiqihong/paxos/network"
)

type Communicator struct {
	myNode   *comm.Node
	msgCoder coder.MsgCoder
	network  network.Network
	actors   []comm.Actor
}

func (c *Communicator) SendMessage(node *comm.Node, msg comm.Message) error {
	msgBytes, err := c.msgCoder.Encode(msg)
	if err != nil {
		return err
	}
	c.network.SendMessage(node.Addr, msgBytes)
	return nil
}

func (c *Communicator) BroadcastMessage(nodes []comm.Node, msg comm.Message, selfMsg bool) error {
	msgBytes, err := c.msgCoder.Encode(msg)
	if err != nil {
		return err
	}

	addrs := make([]string, len(nodes))
	for i, node := range nodes {
		if node.Addr == c.myNode.Addr {
			continue
		}
		addrs[i] = node.Addr
	}

	c.network.BroadcastMessage(addrs, msgBytes)

	if selfMsg {
	}

	return nil
}

func (c *Communicator) OnMessage(msgBytes []byte) error {
	msg, err := c.msgCoder.Decode(msgBytes)
	if err != nil {
		return err
	}

	c.HandleMessage(msg)

	return nil
}

func (c *Communicator) HandleMessage(msg comm.Message) {
	for _, actor := range c.actors {
		if actor.CanHandleMsg(msg) {
			actor.DeliverMsg(msg)
			break
		}
	}
}
