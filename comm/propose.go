package comm

type ProposeResult interface {
	Get() (instanceID uint64, err error)
}

type Proposal interface {
	Cancel()
	GetProposeResultChan() <-chan ProposeResult
	WaitForResult() (instanceID uint64, err error)
}
