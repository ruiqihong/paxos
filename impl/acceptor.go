package impl

import (
	"github.com/ruiqihong/paxos/comm"
	"github.com/ruiqihong/paxos/store"
)

type acceptorState struct {
	promisedBallot comm.BallotNumber
	acceptedBallot comm.BallotNumber
	acceptedValue  []byte
}

func (s *acceptorState) Reset() {
	s.promisedBallot = comm.BallotNumber(0)
	s.acceptedBallot = comm.BallotNumber(0)
	s.acceptedValue = nil
}

func LoadState(store store.LogStorage) (instanceID uint64, state acceptorState, err error) {
	instanceID, err = store.GetMaxInstanceID()
	if err != nil {
		return
	}

	s, err := store.ReadState(instanceID)
	if err != nil {
		return
	}

	state.promisedBallot = comm.BallotNumber(s.PromisedBallot)
	state.acceptedBallot = comm.BallotNumber(s.AcceptedBallot)
	state.acceptedValue = s.AcceptedValue
	return
}

func PersistState(logstore store.LogStorage, instanceID uint64, state acceptorState) error {
	return logstore.WriteState(instanceID, &store.PaxosState{
		instanceID,
		uint64(state.promisedBallot),
		uint64(state.acceptedBallot),
		state.acceptedValue,
	})
}

type AcceptorPrivateConfig interface {
}

type AcceptorConfig interface {
	comm.Nodes
	AcceptorPrivateConfig
}

type Acceptor struct {
	config       AcceptorConfig
	nodes        []comm.Node
	communicator *Communicator
	store        store.LogStorage
	state        acceptorState
	msgCh        chan comm.Message
	newCh        chan chan bool
	instanceID   uint64
}

func NewAcceptor(config AcceptorConfig,
	communicator *Communicator,
	store store.LogStorage) *Acceptor {
	return &Acceptor{
		config:       config,
		communicator: communicator,
		store:        store,
		msgCh:        make(chan comm.Message, 100),
		newCh:        make(chan chan bool, 1),
	}
}

func (ac *Acceptor) NewInstance() <-chan bool {
	done := make(chan bool, 1)
	ac.newCh <- done
	return done
}

func (ac *Acceptor) CanHandleMsg(msg comm.Message) bool {
	switch msg.(type) {
	case *comm.PrepareMsg,
		*comm.AcceptMsg:
		return true
	}
	return false
}

func (ac *Acceptor) DeliverMsg(msg comm.Message) {
	select {
	case ac.msgCh <- msg:
	default:
	}
}

func (ac *Acceptor) Run() error {
	var err error
	ac.instanceID, ac.state, err = LoadState(ac.store)
	if err != nil {
		return err
	}

	for {
		select {
		case msg := <-ac.msgCh:
			err := ac.handleMsg(msg)
			if err != nil {
				return err
			}
		case done := <-ac.newCh:
			ac.instanceID++
			ac.state.Reset()
			done <- true
		}
	}

	return nil
}

func (ac *Acceptor) handleMsg(msg comm.Message) error {
	switch msg.(type) {
	case *comm.PrepareMsg:
		return ac.onPrepare(msg.(*comm.PrepareMsg))
	case *comm.AcceptMsg:
		return ac.onAccept(msg.(*comm.AcceptMsg))
	}
	return nil
}

func (ac *Acceptor) onPrepare(msg *comm.PrepareMsg) error {
	reply := &comm.PrepareReplyMsg{
		NodeID:     ac.config.GetMyNode().NodeID,
		InstanceID: ac.instanceID,
		ProposalID: msg.ProposalID,
	}

	ballot := comm.NewBallotNumber(msg.NodeID, msg.ProposalID)
	if ballot.CompareTo(ac.state.promisedBallot) >= 0 {
		reply.PreAcceptNodeID = ac.state.acceptedBallot.GetNodeID()
		reply.PreAcceptProposalID = ac.state.acceptedBallot.GetProposalID()

		if ac.state.acceptedBallot.GetProposalID() > 0 {
			reply.PreAcceptValue = ac.state.acceptedValue
		}

		ac.state.promisedBallot = ballot
		err := PersistState(ac.store, ac.instanceID, ac.state)
		if err != nil {
			return err
		}
	} else {
		reply.RejectByPromiseID = ac.state.promisedBallot.GetProposalID()
	}

	ac.communicator.SendMessage(comm.GetNodeByID(ac.nodes, msg.NodeID), reply)

	return nil
}

func (ac *Acceptor) onAccept(msg *comm.AcceptMsg) error {
	reply := &comm.AcceptReplyMsg{
		NodeID:     ac.config.GetMyNode().NodeID,
		InstanceID: ac.instanceID,
		ProposalID: msg.ProposalID,
	}

	ballot := comm.NewBallotNumber(msg.NodeID, msg.ProposalID)
	if ballot.CompareTo(ac.state.promisedBallot) >= 0 {
		ac.state.promisedBallot = ballot
		ac.state.acceptedBallot = ballot
		ac.state.acceptedValue = msg.Value
		err := PersistState(ac.store, ac.instanceID, ac.state)
		if err != nil {
			return err
		}
	} else {
		reply.RejectByPromiseID = ac.state.promisedBallot.GetProposalID()
	}

	ac.communicator.SendMessage(comm.GetNodeByID(ac.nodes, msg.NodeID), reply)

	return nil
}
