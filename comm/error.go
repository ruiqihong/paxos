package comm

type Error struct {
	Code int
	Msg  string
}

func (e Error) Error() string {
	return e.Msg
}

const (
	ERR_PROPOSE_TIMEOUT = 101
)
