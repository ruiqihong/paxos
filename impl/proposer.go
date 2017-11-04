package impl

import (
	"time"

	"github.com/ruiqihong/paxos/comm"
)

type ProposerConfig interface {
	comm.Nodes
	ProposerPrivateConfig
}

type ProposerPrivateConfig interface {
	GetPrepareTimeout() time.Duration
	GetAcceptTimeout() time.Duration
}

type proposerState struct {
	proposal                    *proposal
	highestOtherProposalID      uint32
	highestOtherPreAcceptBallot comm.BallotNumber
	proposalID                  uint32
	value                       []byte
}

func (s *proposerState) SetOtherProposalID(proposalID uint32) {
	if proposalID > s.highestOtherProposalID {
		s.highestOtherProposalID = proposalID
	}
}

func (s *proposerState) AddPreAcceptValue(ballot comm.BallotNumber, value []byte) {
	if ballot.IsNull() {
		return
	}

	if ballot.CompareTo(s.highestOtherPreAcceptBallot) > 0 {
		s.highestOtherPreAcceptBallot = ballot
		s.value = value
	}
}

const (
	PROPOSER_STATUS_WAITING    = 0
	PROPOSER_STATUS_PREPARING  = 1
	PROPOSER_STATUS_ACCEPTING  = 2
	PROPOSER_STATUS_RESTARTING = 3
)

type Proposer struct {
	config       ProposerConfig
	communicator *Communicator
	learner      Learner

	proposeCh chan *proposal
	msgCh     chan comm.Message
	newCh     chan chan bool

	state  proposerState
	status uint8

	timer *time.Timer

	msgCounter         MsgCounter
	wasRejectBySomeone bool

	instanceID uint64
}

func NewProposer(config ProposerConfig,
	communicator *Communicator,
	learner Learner) *Proposer {
	proposer := &Proposer{
		config:       config,
		communicator: communicator,
		learner:      learner,
		proposeCh:    make(chan *proposal),
		msgCh:        make(chan comm.Message),
		newCh:        make(chan chan bool),
		msgCounter:   MsgCounter{MsgCounterConfig: config},
		timer:        time.NewTimer(0),
	}

	proposer.msgCounter.StartNewRound()

	return proposer
}

func (p *Proposer) NewInstance() <-chan bool {
	done := make(chan bool, 1)
	p.newCh <- done
	return done
}

func (p *Proposer) Propose(value []byte) comm.Proposal {
	proposal := newProposal(value)
	select {
	case p.proposeCh <- proposal:
	case <-proposal.cancelCh:
	}
	return proposal
}

func (p *Proposer) CanHandleMsg(msg comm.Message) bool {
	switch msg.(type) {
	case *comm.PrepareReplyMsg:
		return true
	}
	return false
}

func (p *Proposer) DeliverMsg(msg comm.Message) {
	select {
	case p.msgCh <- msg:
	default:
	}
}

func (p *Proposer) Run() error {
	for {
		err := p.processEvent()
		if err != nil {
			return err
		}
	}
}

func (p *Proposer) resetTimer(d time.Duration) {
	if !p.timer.Stop() {
		<-p.timer.C
	}
	p.timer.Reset(d)
}

func (p *Proposer) processEvent() error {
	var err error

	if p.status == PROPOSER_STATUS_WAITING {
		select {
		case proposal := <-p.proposeCh:
			err = p.newProposal(proposal)
		case _ = <-p.msgCh:
			// ignore msg, if waiting
		case done := <-p.newCh:
			p.instanceID++
			done <- true
		}
	} else {
		select {
		case msg := <-p.msgCh:
			err = p.handleMsg(msg)
		case <-p.timer.C:
			err = p.handleTimeout()
		case done := <-p.newCh:
			p.instanceID++
			done <- true
		}
	}

	return err
}

func (p *Proposer) handleMsg(msg comm.Message) error {
	var err error
	switch msg.(type) {
	case *comm.PrepareReplyMsg:
		err = p.onPrepareReply(msg.(*comm.PrepareReplyMsg))
	case *comm.AcceptReplyMsg:
		err = p.onAcceptReply(msg.(*comm.AcceptReplyMsg))
	}
	return err
}

func (p *Proposer) handleTimeout() error {
	switch p.status {
	case PROPOSER_STATUS_PREPARING:
		return p.onPrepareTimeout()
	case PROPOSER_STATUS_ACCEPTING:
		return p.onAcceptTimeout()
	case PROPOSER_STATUS_RESTARTING:
		return p.onRestartTimeout()
	}

	return nil
}

func (p *Proposer) newProposal(proposal *proposal) error {
	p.state.proposal = proposal
	return p.prepare(false)
}

func (p *Proposer) prepare(needNewBallot bool) error {
	proposalID := uint32(0)
	if needNewBallot {
		if p.state.proposalID > p.state.highestOtherProposalID {
			proposalID = p.state.proposalID
		} else {
			proposalID = p.state.highestOtherProposalID
		}
		proposalID++
	}

	proposalID++
	p.state.value = p.state.proposal.value
	p.state.proposalID = proposalID

	p.wasRejectBySomeone = false
	p.msgCounter.StartNewRound()

	msg := comm.PrepareMsg{
		NodeID:     p.config.GetMyNode().NodeID,
		InstanceID: p.instanceID,
		ProposalID: proposalID,
	}

	p.status = PROPOSER_STATUS_PREPARING

	p.communicator.BroadcastMessage(p.config.GetNodes(), msg, true)

	return nil
}

func (p *Proposer) onPrepareReply(msg *comm.PrepareReplyMsg) error {
	if p.status != PROPOSER_STATUS_PREPARING {
		return nil
	}

	if msg.ProposalID != p.state.proposalID {
		return nil
	}

	p.msgCounter.AddReceiveNode(msg.NodeID)

	if msg.RejectByPromiseID == 0 {
		ballot := comm.NewBallotNumber(msg.PreAcceptNodeID, msg.PreAcceptProposalID)
		p.msgCounter.AddPromiseOrAcceptNode(msg.NodeID)
		p.state.AddPreAcceptValue(ballot, msg.PreAcceptValue)
	} else {
		p.msgCounter.AddRejectNode(msg.NodeID)
		p.wasRejectBySomeone = true
		p.state.SetOtherProposalID(msg.RejectByPromiseID)
	}

	if p.msgCounter.IsPassedOnThisRound() {
		p.accept()
	} else if p.msgCounter.IsRejectOnThisRound() ||
		p.msgCounter.IsAllReceiveOnThisRound() {
		p.resetTimer(30 * time.Millisecond)
	}

	return nil
}

func (p *Proposer) onPrepareTimeout() error {
	p.status = PROPOSER_STATUS_RESTARTING
	p.resetTimer(30 * time.Millisecond)

	return nil
}

func (p *Proposer) accept() {
	p.status = PROPOSER_STATUS_ACCEPTING

	msg := &comm.AcceptMsg{
		NodeID:     p.config.GetMyNode().NodeID,
		ProposalID: p.state.proposalID,
		InstanceID: p.instanceID,
		Value:      p.state.value,
	}

	p.msgCounter.StartNewRound()
	p.resetTimer(p.config.GetAcceptTimeout())

	p.communicator.BroadcastMessage(p.config.GetNodes(), msg, true)
}

func (p *Proposer) onAcceptReply(msg *comm.AcceptReplyMsg) error {
	if p.status != PROPOSER_STATUS_ACCEPTING {
		return nil
	}

	if msg.ProposalID != p.state.proposalID {
		return nil
	}

	p.msgCounter.AddReceiveNode(msg.NodeID)

	if msg.RejectByPromiseID == 0 {
		p.msgCounter.AddPromiseOrAcceptNode(msg.NodeID)
	} else {
		p.msgCounter.AddRejectNode(msg.NodeID)
		p.wasRejectBySomeone = true
		p.state.SetOtherProposalID(msg.RejectByPromiseID)
	}

	if p.msgCounter.IsPassedOnThisRound() {
		p.status = PROPOSER_STATUS_WAITING
		p.learner.LearnNewValue(p.instanceID, p.state.proposalID, p.state.value)
	} else if p.msgCounter.IsRejectOnThisRound() ||
		p.msgCounter.IsAllReceiveOnThisRound() {
		p.resetTimer(30 * time.Millisecond)
	}
	return nil
}

func (p *Proposer) onAcceptTimeout() error {
	p.status = PROPOSER_STATUS_RESTARTING
	p.resetTimer(30 * time.Millisecond)

	return nil
}

func (p *Proposer) onRestartTimeout() error {
	return p.prepare(p.wasRejectBySomeone)
}

type proposeResult struct {
	InstanceID uint64
	Err        comm.Error
}

func (r proposeResult) Get() (uint64, error) {
	return r.InstanceID, r.Err
}

type proposal struct {
	value    []byte
	cancelCh chan bool
	resCh    chan comm.ProposeResult
}

func newProposal(value []byte) *proposal {
	return &proposal{
		value:    value,
		cancelCh: make(chan bool, 1),
		resCh:    make(chan comm.ProposeResult, 1),
	}
}

func (p *proposal) Cancel() {
	p.cancelCh <- true
}

func (p *proposal) WaitForResult() (instanceID uint64, err error) {
	result := <-p.resCh
	return result.Get()
}

func (p *proposal) GetProposeResultChan() <-chan comm.ProposeResult {
	return p.resCh
}

