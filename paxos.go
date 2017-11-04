package paxos

import (
	"fmt"

	"github.com/ruiqihong/paxos/comm"
	"github.com/ruiqihong/paxos/impl"
	"github.com/ruiqihong/paxos/log"
	"github.com/ruiqihong/paxos/plugin"
)

type Paxos struct {
	config Config
	plugin plugin.Plugin
	groups []*impl.Instance
}

func New(config Config, plug plugin.Plugin) *Paxos {
	if plug == nil {
		plug = plugin.NewDefaultPlugin(config.GetDataPath())
	}
	return &Paxos{
		config: config,
		plugin: plug,
	}
}

func (p *Paxos) Propose(groupID uint, data []byte) comm.Proposal {
	return p.groups[groupID].Propose(data)
}

func (p *Paxos) Run() error {
	p.groups = make([]*impl.Instance, p.config.GetGroupCount())
	errCh := make(chan InstanceErr, 1)

	for i := uint32(0); i < p.config.GetGroupCount(); i++ {
		p.groups[i] = impl.NewInstance(i, p.config, p.plugin)
		go func(i uint32) {
			err := p.groups[i].Run()
			errCh <- InstanceErr{GroupID: i, Err: err}
		}(i)
	}
	err := <-errCh
	log.With(log.F{"err": err}).Err("paxos exit")

	return err
}

type InstanceErr struct {
	GroupID uint32
	Err     error
}

func (e InstanceErr) Error() string {
	return fmt.Sprintf("groupid %s err %+v",
		e.GroupID,
		e.Err)
}
