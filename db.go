package lazydb

import (
	"lazydb/ds"
	"lazydb/logfile"
	"sync"
)

type (
	LazyDB struct {
		cfg              *DBConfig
		index            *ds.ConcurrentMap[string]
		strIndex         *strIndex
		hashIndex        *hashIndex
		fidsMap          map[valueType]*MutexFids
		activeLogFileMap map[valueType]*MutexLogFile
		archivedLogFile  map[valueType]*ds.ConcurrentMap[uint32] // [uint32]*MutexLogFile
		mu               sync.RWMutex
	}

	MutexFids struct {
		fids []uint32
		mu   sync.RWMutex
	}

	MutexLogFile struct {
		lf *logfile.LogFile
		mu sync.RWMutex
	}

	valueType uint8

	strIndex struct {
		mu      *sync.RWMutex
		idxTree *ds.AdaptiveRadixTree
	}

	hashIndex struct {
		mu    *sync.RWMutex
		trees map[string]*ds.AdaptiveRadixTree
	}

	Value struct {
		value     []byte
		vType     valueType
		fid       uint32
		offset    int64
		entrySize int
		expiredAt int64
	}

	// 写LogFile之后返回位置信息的结构体
	ValuePos struct {
		fid       uint32
		offset    int64
		entrySize int
	}
)

const (
	valueTypeString valueType = iota
	valueTypeList
	valueTypeHash
	valueTypeSet
	valueTypeZSet

	logFileTypeNum = 5

	encodeHeaderSize = 10
)

func Open(cfg DBConfig) (*LazyDB, error) {
	return nil, nil
}

// Sync flush the buffer into stable storage.
func (db *LazyDB) Sync() error {
	return nil
}

// Close db
func (db *LazyDB) Close() error {
	return nil
}

func (db *LazyDB) Merge(typ valueType, targetFid uint32) error {
	return nil
}
