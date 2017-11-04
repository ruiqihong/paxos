package impl

type Learner interface {
	LearnNewValue(instanceID uint64, proposalID uint32, value []byte) error
}

type DefaultLearner struct {
}

func NewLearner() *DefaultLearner {
	return &DefaultLearner{}
}

func (*DefaultLearner) LearnNewValue(instanceID uint64, proposalID uint32, value []byte) error {
	return nil
}
