package skl

import (
	"log"
	"sync/atomic"
	"unsafe"

	"github.com/crazyfrankie/nyxdb/internal/kv"
)

const (
	offsetSize = int(unsafe.Sizeof(uint32(0)))

	// nodeAlign is used for memory alignment, whether on a 32-bit or 64-bit machine,
	// so that the node.value field is 64-bit aligned.
	// This is necessary because node.getValue uses atomic.LoadUint64,
	// which expects its input to be 64-bit.
	nodeAlign = int(unsafe.Sizeof(uint64(0))) - 1
)

type Arena struct {
	cnt atomic.Uint32
	buf []byte
}

func newArena(maxSize int64) *Arena {
	// Badger design here, reserving index 0 to prevent data from being stored at offset 0.
	// "Don't store data at position 0 in order to reserve offset=0 as a kind of nil pointer."
	ar := &Arena{buf: make([]byte, maxSize)}
	ar.cnt.Store(1)

	return ar
}

func (a *Arena) size() int64 {
	return int64(a.cnt.Load())
}

// putNode assigns a node in the arena.
// Nodes are aligned to pointer-sized boundary alignments.
// The offset of the node is returned.
func (a *Arena) putNode(height int) uint32 {
	// Calculate the amount that won't be used and truncate it,
	// since height must be less than maxHeight.
	unusedSize := (maxHeight - height) * offsetSize
	total := uint32(MaxNodeSize - unusedSize + nodeAlign)
	n := a.cnt.Add(total)
	if int(n) > len(a.buf) {
		log.Fatalf("Arena is too small, toWrite:%d, newTotal:%d, limit:%d", total, n, len(a.buf))
	}
	// returns the offset after alignment.
	offset := (n - total + uint32(nodeAlign)) &^ uint32(nodeAlign)

	return offset
}

func (a *Arena) putKey(key []byte) uint32 {
	total := uint32(len(key))
	n := a.cnt.Add(total)
	if int(n) > len(a.buf) {
		log.Fatalf("Arena is too small, toWrite:%d, newTotal:%d, limit:%d", total, n, len(a.buf))
	}
	offset := n - total
	copy(a.buf[offset:n], key)

	return offset
}

// Put will *copy* val into arena. To make better use of this, reuse your input
// val buffer. Returns an offset into buf. User is responsible for remembering
// size of val. We could also store this size inside arena but the encoding and
// decoding will incur some overhead.
func (a *Arena) putVal(val kv.Value) uint32 {
	total := val.EncodedSize()
	n := a.cnt.Add(total)

	if int(n) > len(a.buf) {
		log.Fatalf("Arena is too small, toWrite:%d, newTotal:%d, limit:%d", total, n, len(a.buf))
	}
	offset := n - total
	val.Encode(a.buf[offset:])

	return offset
}

// getNode returns a pointer to the node located at offset. If the offset is zero,
// then the nil node pointer is returned.
func (a *Arena) getNode(offset uint32) *node {
	if offset == 0 {
		return nil
	}

	return (*node)(unsafe.Pointer(&a.buf[offset]))
}

// getKey returns byte slice at offset.
func (a *Arena) getKey(offset uint32, size uint16) []byte {
	return a.buf[offset : offset+uint32(size)]
}

// getVal returns byte slice at offset.
func (a *Arena) getVal(offset, size uint32) (res kv.Value) {
	res.Decode(a.buf[offset : offset+size])
	return
}

// getNodeOffset returns the offset of node in the arena. If the node pointer is
// nil, then the zero offset is returned.
func (a *Arena) getNodeOffset(n *node) uint32 {
	if n == nil {
		return 0
	}

	return uint32(uintptr(unsafe.Pointer(n)) - uintptr(unsafe.Pointer(&a.buf[0])))
}
