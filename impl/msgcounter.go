package impl

import (
	"github.com/ruiqihong/paxos/comm"
)

type MsgCounterConfig interface {
	GetNodes() []comm.Node
	GetMajorityCount() uint
}

type MsgCounter struct {
	ReceiveMsgNodeID      map[uint32]bool
	RejectMsgNodeID       map[uint32]bool
	PromiseOrAcceptNodeID map[uint32]bool
	MsgCounterConfig
}

func (mc *MsgCounter) StartNewRound() {
	mc.ReceiveMsgNodeID = make(map[uint32]bool)
	mc.RejectMsgNodeID = make(map[uint32]bool)
	mc.PromiseOrAcceptNodeID = make(map[uint32]bool)
}

func (mc *MsgCounter) AddReceiveNode(nodeID uint32) {
	mc.ReceiveMsgNodeID[nodeID] = true
}

func (mc *MsgCounter) AddRejectNode(nodeID uint32) {
	mc.RejectMsgNodeID[nodeID] = true
}

func (mc *MsgCounter) AddPromiseOrAcceptNode(nodeID uint32) {
	mc.PromiseOrAcceptNodeID[nodeID] = true
}

func (mc *MsgCounter) IsPassedOnThisRound() bool {
	return uint(len(mc.PromiseOrAcceptNodeID)) >= mc.GetMajorityCount()
}

func (mc *MsgCounter) IsRejectOnThisRound() bool {
	return uint(len(mc.RejectMsgNodeID)) >= mc.GetMajorityCount()
}

func (mc *MsgCounter) IsAllReceiveOnThisRound() bool {
	return len(mc.ReceiveMsgNodeID) == len(mc.GetNodes())
}
