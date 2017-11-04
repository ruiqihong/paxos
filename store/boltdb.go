package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/boltdb/bolt"
)

type BoltDBStorageFactory struct {
	dataPath string
}

func NewBoltDBStorageFactory(dataPath string) *BoltDBStorageFactory {
	return &BoltDBStorageFactory{dataPath: dataPath}
}

func (fac *BoltDBStorageFactory) GetLogStorage(groupID uint32) LogStorage {
	return NewBoltDBStorage(fac.dataPath, groupID)
}

type BoltDBStorage struct {
	dataPath string
	groupID  uint32
	db       *bolt.DB
}

func NewBoltDBStorage(dataPath string, groupID uint32) *BoltDBStorage {
	return &BoltDBStorage{dataPath: dataPath, groupID: groupID}
}

func (s *BoltDBStorage) Init() (err error) {
	s.db, err = bolt.Open(fmt.Sprintf("%s/paxoslog_%u", s.dataPath, s.groupID), 0600, nil)
	return
}

func (s *BoltDBStorage) Close() error {
	return s.db.Close()
}

func (s *BoltDBStorage) WriteState(instanceID uint64, state *PaxosState) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		var err error
		metaBucket := tx.Bucket([]byte("Meta"))
		value := metaBucket.Get([]byte("MaxInstanceID"))

		var maxInstanceID uint64
		maxInstanceID, err = strconv.ParseUint(string(value), 16, 64)
		if err != nil {
			return err
		}

		if instanceID < maxInstanceID {
			return errors.New("invalid instanceID")
		}

		instanceBucket := tx.Bucket([]byte("Instance"))
		value, err = json.Marshal(state)
		if err != nil {
			return err
		}
		err = instanceBucket.Put([]byte(strconv.FormatUint(instanceID, 16)), value)
		if err != nil {
			return err
		}

		err = metaBucket.Put([]byte("MaxInstanceID"), []byte(strconv.FormatUint(instanceID, 16)))

		return err
	})
}

func (s *BoltDBStorage) ReadState(instanceID uint64) (state *PaxosState, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Instance"))
		value := b.Get([]byte(strconv.FormatUint(instanceID, 16)))
		return json.Unmarshal(value, state)
	})
	return
}

func (s *BoltDBStorage) GetMaxInstanceID() (instanceID uint64, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte("Meta"))
		value := b.Get([]byte("MaxInstanceID"))
		instanceID, err = strconv.ParseUint(string(value), 16, 64)
		return err
	})
	return
}
