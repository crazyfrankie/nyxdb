package iterator

import (
	"github.com/crazyfrankie/nyxdb/internal/kv"
)

// Iterator is primarily used for memtable traversal,
// for swiping, range queries, and transactions
type Iterator interface {
	Next()
	Rewind()
	Seek(key []byte)
	Key() []byte
	Value() kv.Value
	Valid() bool

	// Close All iterators should be closed so that file garbage collection works.
	Close() error
}
