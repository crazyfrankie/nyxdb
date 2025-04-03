package skl

import (
	"log"
	"math/rand"
	"sync/atomic"
	"unsafe"

	"github.com/crazyfrankie/nyxdb/internal/kv"
	"github.com/crazyfrankie/nyxdb/internal/util"
)

const (
	maxHeight = 20

	// MaxNodeSize is the memory footprint of a node of maximum height.
	MaxNodeSize = int(unsafe.Sizeof(node{}))
)

type node struct {
	// val is divided into parts in an Uint64 unit to facilitate atomic operations
	// offset is:
	// value offset (bits 0-31)
	// value size (bits 32-63)
	value atomic.Uint64

	keyOffset uint32
	keySize   uint16

	// height of the next
	height uint16

	// The next array represents the next element a node points to at each level.
	// Since most nodes don't need a full height in the next array.
	// since the probability of each layer decreases exponentially.
	// Therefore, these nodes are truncated during allocation to avoid containing unneeded next elements
	next [maxHeight]atomic.Uint32
}

type SkipList struct {
	head   *node
	height atomic.Int32
	ref    atomic.Int32
	arena  *Arena
}

// IncrRef increases the refcount
func (s *SkipList) IncrRef() {
	s.ref.Add(1)
}

// DecrRef decrements the refcount, deallocating the SkipList when done using it
func (s *SkipList) DecrRef() {
	newRef := s.ref.Add(-1)
	if newRef > 0 {
		return
	}

	// Indicate we are closed. Good for testing.  Also, lets GC reclaim memory. Race condition
	// here would suggest we are accessing SkipList when we are supposed to have no reference!
	s.arena = nil
	// Since the head references the arena's buf, as long as the head is kept around
	// GC can't release the buf.
	s.head = nil
}

func encodeValue(valueOffset uint32, valueSize uint32) uint64 {
	return uint64(valueSize)<<32 | uint64(valueOffset)
}

func decodeValue(value uint64) (valOffset uint32, valSize uint32) {
	valOffset = uint32(value)
	valSize = uint32(value >> 32)
	return
}

func newNode(a *Arena, key []byte, val kv.Value, height int) *node {
	// The base level is already allocated in the node struct.
	offset := a.putNode(height)
	n := a.getNode(offset)
	n.keyOffset = a.putKey(key)
	n.keySize = uint16(len(key))
	n.height = uint16(height)
	n.value.Store(encodeValue(a.putVal(val), val.EncodedSize()))
	return n
}

func NewSkipList(arenaSize int64) *SkipList {
	arena := newArena(arenaSize)
	head := newNode(arena, nil, kv.Value{}, maxHeight)
	skl := &SkipList{head: head, arena: arena}
	skl.height.Store(1)
	skl.ref.Add(1)
	return skl
}

// getValue returns valueOffset and valSize
func (n *node) getValue() (uint32, uint32) {
	value := n.value.Load()
	return decodeValue(value)
}

// setValue stores the given val in the node and arena.
func (n *node) setValue(a *Arena, val kv.Value) {
	valueOffset := a.putVal(val)
	value := encodeValue(valueOffset, val.EncodedSize())
	n.value.Store(value)
}

// getNextOffset returns the offset of the node with the given height in the next array.
func (n *node) getNextOffset(height int) uint32 {
	return n.next[height].Load()
}

// key returns the key for the node from the arena.
func (n *node) key(a *Arena) []byte {
	return a.getKey(n.keyOffset, n.keySize)
}

// getHeight returns the current height of the SkipList.
func (s *SkipList) getHeight() int32 {
	return s.height.Load()
}

// getNext returns the node with the corresponding height in the next array for a given node.
func (s *SkipList) getNext(n *node, height int) *node {
	return s.arena.getNode(n.getNextOffset(height))
}

// findNear finds the node near to key.
// If less=true, it finds rightmost node such that node.key < key (if allowEqual=false) or node.key <= key (if allowEqual=true).
// If less=false, it finds leftmost node such that node.key > key (if allowEqual=false) or node.key >= key (if allowEqual=true).
// Returns the node found. The bool returned is true if the node has key equal to given key.
func (s *SkipList) findNear(key []byte, less bool, allowEqual bool) (*node, bool) {
	curr := s.head
	level := int(s.getHeight() - 1)
	for {
		// Assume curr.key < key.
		next := s.getNext(curr, level)
		if next == nil {
			// curr.key < key < END OF LIST
			if level > 0 {
				level--
				continue
			}
			// Level=0. Cannot descend further. At this point we continue to look for.
			if !less {
				return nil, false
			}
			if curr == s.head {
				return nil, false
			}
			return curr, false
		}

		nextKey := next.key(s.arena)
		cmp := util.CompareKeys(key, nextKey)
		if cmp > 0 {
			// curr.key < next.key < key. We can continue to move right.
			curr = next
			continue
		}
		if cmp == 0 {
			// curr.key < key == next.key.
			if allowEqual {
				return next, true
			}
			if !less {
				// We want >, so go to base level to grab the next bigger note.
				return s.getNext(next, 0), false
			}
			// We want <. If not base level, we should go closer in the next level.
			if level > 0 {
				level--
				continue
			}
			// On base level. Return curr.
			if curr == s.head {
				return nil, false
			}
			return curr, false
		}
		// cmp < 0. In other words, curr.key < key < next.key
		if level > 0 {
			level--
			continue
		}
		if !less {
			return next, false
		}
		if curr == s.head {
			return nil, false
		}
		return curr, false
	}
}

// findLast is used to locate the last node of the current SkipList
func (s *SkipList) findLast() *node {
	curr := s.head
	level := int(s.getHeight()) - 1
	for {
		next := s.getNext(curr, level)
		if next != nil {
			curr = next
			continue
		}
		if level == 0 {
			if curr == s.head {
				return nil
			}
			return curr
		}
		level--
	}
}

// Get gets the value associated with the key.
// It returns a valid value if it finds equal or earlier version of the same key.
func (s *SkipList) Get(key []byte) kv.Value {
	n, _ := s.findNear(key, false, true) // findGreaterOrEqual
	if n == nil {
		return kv.Value{}
	}

	nextKey := s.arena.getKey(n.keyOffset, n.keySize)
	if !util.SameKey(key, nextKey) {
		return kv.Value{}
	}

	valOffset, valSize := n.getValue()
	vs := s.arena.getVal(valOffset, valSize)
	vs.Version = util.ParseTs(key)

	return vs
}

// findSpliceForLevel Finds the insertion position in the given hierarchical level
func (s *SkipList) findSpliceForLevel(key []byte, before *node, level int) (*node, *node) {
	for {
		// assume before.key < key
		next := s.getNext(before, level)
		if next == nil {
			return before, next
		}
		nextKey := next.key(s.arena)
		cmp := util.CompareKeys(key, nextKey)
		if cmp == 0 {
			// Equal case
			// just update val
			return next, next
		}
		if cmp < 0 {
			// before.key < key < next.key. We are done for this level.
			// insert between before and next
			return before, next
		}
		// continue find
		before = next
	}
}

// Put inserts the key-value pair.
func (s *SkipList) Put(key []byte, val kv.Value) {
	currHeight := s.getHeight()
	var prev [maxHeight + 1]*node
	var next [maxHeight + 1]*node
	prev[currHeight] = s.head
	next[currHeight] = nil

	for i := int(currHeight) - 1; i >= 0; i-- {
		prev[i], next[i] = s.findSpliceForLevel(key, prev[i+1], i)
		if prev[i] == next[i] {
			prev[i].setValue(s.arena, val)
			return
		}
	}

	height := s.RandomLevel()
	newNode := newNode(s.arena, key, val, height)

	// Try to increase height through CAS
	currHeight = s.getHeight()
	for height > int(currHeight) {
		if s.height.CompareAndSwap(currHeight, int32(height)) {
			// Successfully increased SkipList.height.
			break
		}
		currHeight = s.getHeight()
	}

	// Add nodes from the base level
	for i := 0; i < height; i++ {
		for {
			if prev[i] == nil {
				if i <= 1 {
					log.Fatalf("Invalid level: %d. This cannot happen in base level.", i)
				}
				// We haven't computed prev, next for this level because height exceeds old currHeight.
				// For these levels, we expect the lists to be sparse, so we can just search from head.
				prev[i], next[i] = s.findSpliceForLevel(key, s.head, i)
				// Someone adds the exact same key before we are able to do so.
				// This can only happen on the base level. But we know we are not on the base level.
				// This doesn't usually happen, but if prev[i] == next[i],
				// there's a problem with the jump table structure (e.g. multiple threads inserting the same key at the same time).
				if prev[i] == next[i] {
					log.Fatalf("prev[i] and next[i] are equal at level %d, which should never happen.", i)
				}
			}
			nextOffset := s.arena.getNodeOffset(next[i])
			newNode.next[i].Store(nextOffset)
			if prev[i].next[i].CompareAndSwap(nextOffset, s.arena.getNodeOffset(newNode)) {
				// Managed to insert newNode between prev[i] and next[i]. Go to the next level.
				break
			}
			// CAS failed, another thread modified prev[i].next[i], need to find it again.
			// Recalculate prev[i] and next[i], but this time continue from prev[i] instead of head
			prev[i], next[i] = s.findSpliceForLevel(key, prev[i], i)
			if prev[i] == next[i] {
				if i != 0 {
					log.Fatalf("Equality can happen only on base level, but found on level %d.", i)
				}
				prev[i].setValue(s.arena, val)
				return
			}
		}
	}
}

// RandomLevel generates a random number of levels
func (s *SkipList) RandomLevel() int {
	level := 1
	for rand.Float32() < 0.5 && level < maxHeight {
		level++
	}
	return level
}

// MemorySize returns the size of the SkipList in terms of how much memory is used within its internal arena.
func (s *SkipList) MemorySize() int64 {
	return s.arena.size()
}

// NewIterator returns a SkipList iterator. You must close the iterator.
func (s *SkipList) NewIterator() *Iterator {
	s.IncrRef()
	return &Iterator{list: s}
}

// Iterator is a bidirectional iterator that can be directly used for range queries, bulk data processing
type Iterator struct {
	list *SkipList
	n    *node
}

func (i *Iterator) Close() error {
	i.list.DecrRef()
	return nil
}

// Valid returns true iff the iterator is positioned at a valid node.
func (i *Iterator) Valid() bool {
	return i.n != nil
}

// Key returns the key at the current node
func (i *Iterator) Key() []byte {
	return i.list.arena.getKey(i.n.keyOffset, i.n.keySize)
}

// Value returns the value at the current node
func (i *Iterator) Value() kv.Value {
	return i.list.arena.getVal(i.n.getValue())
}

// ValueUint64 returns the uint64 value of the current node
func (i *Iterator) ValueUint64() uint64 {
	return i.n.value.Load()
}

// Next moves to the next position.
func (i *Iterator) Next() {
	if !i.Valid() {
		log.Fatalf("the current node is nil, can't move to next node")
	}
	i.n = i.list.getNext(i.n, 0)
}

// Prev moves to the previous position
func (i *Iterator) Prev() {
	if !i.Valid() {
		log.Fatalf("the current node is nil, can't move to prev node")
	}
	i.n, _ = i.list.findNear(i.Key(), true, false)
}

// Seek moves to the first entry with a key >= target
func (i *Iterator) Seek(target []byte) {
	i.n, _ = i.list.findNear(target, false, true) // find >=
}

// SeekPrev moves to the first entry with a key <= target
func (i *Iterator) SeekPrev(target []byte) {
	i.n, _ = i.list.findNear(target, true, true) // find <=
}

// SeekToFirst seeks at the first entry in the list,
// i.e., jumps to the first key in the SkipList, for sequential traversal.
// If the list is not empty, the final state of the iterator is Valid().
func (i *Iterator) SeekToFirst() {
	i.n = i.list.getNext(i.list.head, 0)
}

// SeekToLast seeks at the last entry in the list,
// i.e., jumps to the last key in the SkipList, for inverted traversal.
// If the list is not empty, the final state of the iterator is Valid().
func (i *Iterator) SeekToLast() {
	i.n = i.list.findLast()
}

// UniIterator is a unidirectional memtable iterator that is a simple wrapper around Iterator.
type UniIterator struct {
	iter     *Iterator
	reversed bool
}

// NewUniIterator returns a UniIterator.
func (s *SkipList) NewUniIterator(reversed bool) *UniIterator {
	return &UniIterator{
		iter:     s.NewIterator(),
		reversed: reversed,
	}
}

// Next is unidirectional, so the value of reversed determines
// whether to reverse traversal or not.
func (u *UniIterator) Next() {
	if !u.reversed {
		u.iter.Next()
	} else {
		u.iter.Prev()
	}
}

func (u *UniIterator) Rewind() {
	if !u.reversed {
		u.iter.SeekToFirst()
	} else {
		u.iter.SeekToLast()
	}
}

func (u *UniIterator) Seek(key []byte) {
	if !u.reversed {
		u.iter.Seek(key)
	} else {
		u.iter.SeekPrev(key)
	}
}

func (u *UniIterator) Key() []byte {
	return u.iter.Key()
}

func (u *UniIterator) Value() kv.Value {
	return u.iter.Value()
}

func (u *UniIterator) Valid() bool {
	return u.iter.Valid()
}

func (u *UniIterator) Close() error {
	return u.iter.Close()
}
