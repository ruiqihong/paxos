package plugin

import (
	"github.com/ruiqihong/paxos/coder"
	"github.com/ruiqihong/paxos/network"
	"github.com/ruiqihong/paxos/store"
)

type Plugin interface {
	GetNetwork() network.Network
	GetMsgCoder() coder.MsgCoder
	GetLogStorageFactory() store.LogStorageFactory
}

type DefaultPlugin struct {
	network.Network
	coder.MsgCoder
	store.LogStorageFactory
}

func NewDefaultPlugin(dataPath string) *DefaultPlugin {
	return &DefaultPlugin{
		Network:           network.NewDefaultNetwork(),
		MsgCoder:          coder.DefaultMsgCoder{},
		LogStorageFactory: store.NewBoltDBStorageFactory(dataPath),
	}
}

func (p *DefaultPlugin) GetNetwork() network.Network {
	return p.Network
}

func (p *DefaultPlugin) GetMsgCoder() coder.MsgCoder {
	return p.MsgCoder
}

func (p *DefaultPlugin) GetLogStorageFactory() store.LogStorageFactory {
	return p.LogStorageFactory
}
