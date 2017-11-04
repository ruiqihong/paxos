package coder

import (
	"errors"

	"github.com/golang/protobuf/proto"

	"github.com/ruiqihong/paxos/comm"
)

type MsgCoder interface {
	Encode(comm.Message) ([]byte, error)
	Decode([]byte) (comm.Message, error)
}

const (
	UNKNOWN_MSG       = 0
	PREPARE_MSG       = 1
	PREPARE_REPLY_MSG = 2
	ACCEPT_MSG        = 3
	ACCEPT_REPLY_MSG  = 4
)

type DefaultMsgCoder struct {
}

func (DefaultMsgCoder) GetTypeByMsg(msg comm.Message) uint8 {
	switch msg.(type) {
	case *comm.PrepareMsg:
		return PREPARE_MSG
	case *comm.PrepareReplyMsg:
		return PREPARE_REPLY_MSG
	case *comm.AcceptMsg:
		return ACCEPT_MSG
	case *comm.AcceptReplyMsg:
		return ACCEPT_REPLY_MSG
	}
	return UNKNOWN_MSG
}

func (DefaultMsgCoder) GetMsgByType(t uint8) comm.Message {
	switch t {
	case PREPARE_MSG:
		return &comm.PrepareMsg{}
	case PREPARE_REPLY_MSG:
		return &comm.PrepareReplyMsg{}
	case ACCEPT_MSG:
		return &comm.AcceptMsg{}
	case ACCEPT_REPLY_MSG:
		return &comm.AcceptReplyMsg{}
	}
	return &comm.UnknownMsg{}
}

func (c DefaultMsgCoder) Encode(msg comm.Message) ([]byte, error) {
	m, ok := msg.(proto.Message)
	if !ok {
		return nil, errors.New("can't encode msg")
	}

	data, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 1+len(data))
	buf[0] = c.GetTypeByMsg(msg)
	copy(buf[1:], data)

	return buf, nil
}

func (c DefaultMsgCoder) Decode(b []byte) (msg comm.Message, err error) {
	if len(b) == 0 {
		return nil, errors.New("invalid data len")
	}

	t, ok := c.GetMsgByType(b[0]).(proto.Message)
	if !ok {
		return nil, errors.New("can't get msgtype")
	}

	err = proto.Unmarshal(b, t)
	return msg, err
}
