package nyx

import (
	"bytes"
	"sync"
	"sync/atomic"

	"github.com/dgraph-io/ristretto/v2/z"

	"github.com/crazyfrankie/nyxdb/skl"
)

type memTable struct {
	skl *skl.SkipList
	wal *wal
	opt *option
	buf *bytes.Buffer // cache data to reduce frequent disk writes by WAL
}

func openMemTables(opts ...Option) error {
	o := defaultMemTableOpt
	for _, opt := range opts {
		opt(o)
	}

	return nil
}

func (d *DB) openMemTable(fid int) (*memTable, error) {
	return nil, nil
}

type wal struct {
	mmapFile *z.MmapFile // improve IO performance with mmap
	path     string
	mu       sync.RWMutex
	fd       uint32
	size     atomic.Uint32 // current file size
	writeAt  uint32        // write offset
	mem      *memTable     // store a reference to access opt
}
