package impl

import (
	"github.com/ruiqihong/paxos/coder"
	"github.com/ruiqihong/paxos/comm"
	"github.com/ruiqihong/paxos/log"
	"github.com/ruiqihong/paxos/network"
	"github.com/ruiqihong/paxos/store"
)

type InstanceConfig interface {
	comm.Nodes
	ProposerPrivateConfig
}

type InstancePlugin interface {
	GetNetwork() network.Network
	GetMsgCoder() coder.MsgCoder
	GetLogStorageFactory() store.LogStorageFactory
}

type Instance struct {
	groupID  uint32
	config   InstanceConfig
	network  network.Network
	msgCoder coder.MsgCoder
	store    store.LogStorage

	proposer *Proposer
	acceptor *Acceptor
	actors   []comm.Actor
}

func NewInstance(groupID uint32,
	config InstanceConfig,
	plugin InstancePlugin) *Instance {
	instance := &Instance{
		groupID:  groupID,
		config:   config,
		network:  plugin.GetNetwork(),
		msgCoder: plugin.GetMsgCoder(),
	}

	communicator := &Communicator{
		myNode:   instance.config.GetMyNode(),
		msgCoder: instance.msgCoder,
		network:  instance.network,
		actors:   instance.actors,
	}

	instance.network.SetMessageHandler(communicator)

	learner := NewLearner()

	instance.store = plugin.GetLogStorageFactory().GetLogStorage(groupID)

	instance.proposer = NewProposer(config, communicator, learner)
	instance.acceptor = NewAcceptor(config, communicator, instance.store)

	instance.actors = []comm.Actor{instance.proposer, instance.acceptor}

	return instance
}

func (inst *Instance) Propose(value []byte) comm.Proposal {
	return inst.proposer.Propose(value)
}

func (inst *Instance) NewInstance() {
	<-inst.acceptor.NewInstance()
	inst.proposer.NewInstance()
}

func (inst *Instance) Run() error {
	err := inst.store.Init()
	if err != nil {
		return err
	}

	errCh := make(chan error, 1)

	go func() {
		err := inst.network.Run(inst.config.GetMyNode().Addr)
		errCh <- err
	}()

	for _, actor := range inst.actors {
		go func(actor comm.Actor) {
			err := actor.Run()
			errCh <- err
		}(actor)
	}

	err = <-errCh
	log.With(log.F{"groupID": inst.groupID, "err": err}).Err("instance exit")
	return err
}
