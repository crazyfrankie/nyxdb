package nyx

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/dgraph-io/ristretto/v2/z"

	"github.com/crazyfrankie/nyxdb/skl"
)

const (
	MemTableExt = ".mem"
)

type memTable struct {
	skl *skl.SkipList
	wal *wal
	opt *option
	buf *bytes.Buffer // cache data to reduce frequent disk writing by WAL
}

func (d *DB) openMemTables() error {
	return nil
}

func (d *DB) openMemTable(fid, flags int) (*memTable, error) {
	filepath := d.memTablePath(fid)
	skl := skl.NewSkipList(d.arenaSize())
	mt := &memTable{
		skl: skl,
		opt: d.opt,
		buf: &bytes.Buffer{},
	}
	mt.wal = &wal{
		path:    filepath,
		fid:     uint32(fid),
		writeAt: vlogHeaderSize,
		opt:     d.opt,
	}
	// TODO
	return mt, nil
}

func (d *DB) memTablePath(fid int) string {
	return fmt.Sprintf("%05d%s", fid, MemTableExt)
}

// arenaSize returns default arena size
func (d *DB) arenaSize() int64 {
	return d.opt.MemTableSize + d.opt.maxBatchSize + d.opt.maxBatchCount*int64(skl.MaxNodeSize)
}

type wal struct {
	mmapFile *z.MmapFile // improve IO performance with mmap
	path     string
	mu       sync.RWMutex
	fid      uint32
	size     atomic.Uint32 // current file size
	writeAt  uint32        // write offset
	opt      *option
}
