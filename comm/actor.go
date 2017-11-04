package comm

type Actor interface {
	Run() error
	NewInstance() (done <-chan bool)
	CanHandleMsg(Message) bool
	DeliverMsg(Message)
}
