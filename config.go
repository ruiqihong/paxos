package paxos

import (
	"time"

	"github.com/ruiqihong/paxos/comm"
)

type Config interface {
	comm.Nodes

	GetGroupCount() uint32
	GetDataPath() string
	GetPrepareTimeout() time.Duration
	GetAcceptTimeout() time.Duration
}

type DefaultConfig struct {
	comm.Nodes
	GroupCount     uint32
	DataPath       string
	PrepareTimeout time.Duration
	AcceptTimeout  time.Duration
}

func NewConfig(nodes comm.Nodes) Config {
	return &DefaultConfig{
		Nodes:          nodes,
		GroupCount:     1,
		DataPath:       "data",
		PrepareTimeout: 50 * time.Millisecond,
		AcceptTimeout:  50 * time.Millisecond,
	}
}

func (c *DefaultConfig) GetGroupCount() uint32 {
	return c.GroupCount
}

func (c *DefaultConfig) GetDataPath() string {
	return c.DataPath
}

func (c *DefaultConfig) GetPrepareTimeout() time.Duration {
	return c.PrepareTimeout
}

func (c *DefaultConfig) GetAcceptTimeout() time.Duration {
	return c.AcceptTimeout
}
