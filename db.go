package lazydb

import (
	"errors"
	"lazydb/ds"
	"lazydb/logfile"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type (
	LazyDB struct {
		cfg              *DBConfig
		index            *ds.ConcurrentMap[string]
		strIndex         *strIndex
		hashIndex        *hashIndex
		listIndex        *listIndex
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

	listIndex struct {
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

	initialListSeq = uint32(math.MaxUint32 / 2)
)

var (
	ErrKeyNotFound     = errors.New("key not found")
	ErrLogFileNotExist = errors.New("log file is not exist")
	ErrOpenLogFile     = errors.New("open Log file error")
	ErrWrongIndex      = errors.New("index is out of range")
)

func newStrIndex() *strIndex {
	return &strIndex{idxTree: ds.NewART(), mu: new(sync.RWMutex)}
}

func newHashIndex() *hashIndex {
	return &hashIndex{trees: make(map[string]*ds.AdaptiveRadixTree), mu: new(sync.RWMutex)}
}

func newListIndex() *listIndex {
	return &listIndex{trees: make(map[string]*ds.AdaptiveRadixTree), mu: new(sync.RWMutex)}
}

func Open(cfg DBConfig) (*LazyDB, error) {
	// create the dir path if not exist
	if !util.PathExist(cfg.DBPath) {
		if err := os.MkdirAll(cfg.DBPath, os.ModePerm); err != nil {
			log.Fatalf("Create db directory in %s error: %v", cfg.DBPath, err)
			return nil, err
		}
	}

	db := &LazyDB{
		cfg:              &cfg,
		index:            ds.NewConcurrentMap(int(cfg.HashIndexShardCount)),
		strIndex:         newStrIndex(),
		hashIndex:        newHashIndex(),
		listIndex:        newListIndex(),
		fidsMap:          make(map[valueType]*MutexFids),
		activeLogFileMap: make(map[valueType]*MutexLogFile),
		archivedLogFile:  make(map[valueType]*ds.ConcurrentMap[uint32]),
	}

	for i := 0; i < logFileTypeNum; i++ {
		db.fidsMap[valueType(i)] = &MutexFids{fids: make([]uint32, 0)}
		db.archivedLogFile[valueType(i)] = ds.NewWithCustomShardingFunction[uint32](ds.DefaultShardCount, ds.SimpleSharding)
	}

	if err := db.buildLogFiles(); err != nil {
		log.Fatalf("Build Log Files error: %v", err)
		return nil, err
	}

	//if err := db.buildIndexFromLogFiles(); err != nil {
	//	log.Fatalf("Build Index From Log Files error: %v", err)
	//	return nil, err
	//}

	return db, nil
}

// buildLogFiles Recover archivedLogFile from disk.
// Only run once when program start running.
func (db *LazyDB) buildLogFiles() error {
	fileInfos, err := os.ReadDir(db.cfg.DBPath)
	if err != nil {
		return err
	}
	for _, file := range fileInfos {
		if !strings.HasPrefix(file.Name(), logfile.FilePrefix) {
			continue
		}
		splitInfo := strings.Split(file.Name(), ".")
		if len(splitInfo) != 3 {
			log.Printf("Invalid log file name: %s", file.Name())
			continue
		}
		typ := valueType(logfile.FileTypesMap[splitInfo[1]])
		fid, err := strconv.Atoi(splitInfo[2])
		if err != nil {
			log.Printf("Invalid log file name: %s", file.Name())
			continue
		}
		fids := db.fidsMap[typ]
		fids.fids = append(fids.fids, uint32(fid))
	}

	build := func(typ valueType) {
		mutexFids := db.fidsMap[typ]
		fids := mutexFids.fids
		if len(fids) == 0 {
			return
		}
		// newly created log file has bigger fid
		sort.Slice(fids, func(i, j int) bool {
			return fids[i] < fids[j]
		})
		archivedLogFiles := db.archivedLogFile[typ]
		for i, fid := range fids {
			lf, err := logfile.Open(db.cfg.DBPath, fid, db.cfg.MaxLogFileSize, logfile.FType(typ), db.cfg.IOType)
			if err != nil {
				log.Fatalf("Open Log File error:%v. Type: %v, Fid: %v,", err, typ, fid)
				continue
			}

			// latest one is the active log file
			if i == len(fids)-1 {
				db.activeLogFileMap[typ] = &MutexLogFile{lf: lf}
			} else {
				archivedLogFiles.Set(fid, &MutexLogFile{lf: lf})
			}
		}
	}
	for typ := 0; typ < logFileTypeNum; typ++ {
		build(valueType(typ))
	}
	return nil
}

// Sync flush the buffer into stable storage.
func (db *LazyDB) Sync() error {
	for _, mlf := range db.activeLogFileMap {
		mlf.mu.Lock()
		if err := mlf.lf.Sync(); err != nil {
			return err
		}
		mlf.mu.Unlock()
	}
	return nil
}

// Close db
func (db *LazyDB) Close() error {
	return nil
}

func (db *LazyDB) Merge(typ valueType, targetFid uint32) error {
	return nil
}
