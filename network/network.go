package network

type MessageHandler interface {
	OnMessage(msg []byte) error
}

type Network interface {
	Run(addr string) error
	SendMessage(addr string, msg []byte)
	SetMessageHandler(handler MessageHandler)
}
