package comm

type BallotNumber uint64

func NewBallotNumber(nodeID uint32, proposalID uint32) BallotNumber {
	return BallotNumber(uint64(proposalID<<32) | uint64(nodeID))
}

func (n BallotNumber) CompareTo(o BallotNumber) int {
	if n < o {
		return -1
	} else if n > 0 {
		return 1
	}

	return 0
}

func (n BallotNumber) GetProposalID() uint32 {
	return uint32(n >> 32)
}

func (n BallotNumber) GetNodeID() uint32 {
	return uint32(n)
}

func (n BallotNumber) IsNull() bool {
	return n == 0
}
