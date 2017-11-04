package store

type LogStorageFactory interface {
	GetLogStorage(groupID uint32) LogStorage
}

type LogStorage interface {
	Init() error
	Close() error

	WriteState(instanceID uint64, state *PaxosState) error
	ReadState(instanceID uint64) (*PaxosState, error)
	GetMaxInstanceID() (uint64, error)
}
