package skl

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/crazyfrankie/nyxdb/internal/kv"
	"github.com/crazyfrankie/nyxdb/internal/util"
)

const arenaSize = 1 << 20

func newValue(v int) []byte {
	return []byte(fmt.Sprintf("%05d", v))
}

func TestBasicWay(t *testing.T) {
	l := NewSkipList(arenaSize)
	val1 := newValue(42)
	val2 := newValue(52)
	val3 := newValue(62)
	val4 := newValue(72)
	val5 := []byte(fmt.Sprintf("%0102400d", 1)) // Have size 100 KB which is > math.MaxUint16.

	// Try inserting values.
	// Somehow require.Nil doesn't work when checking for unsafe.Pointer(nil).
	l.Put(util.KeyWithTs([]byte("key1"), 0), kv.Value{Value: val1, Meta: 55, UserMeta: 0})
	l.Put(util.KeyWithTs([]byte("key2"), 2), kv.Value{Value: val2, Meta: 56, UserMeta: 0})
	l.Put(util.KeyWithTs([]byte("key3"), 0), kv.Value{Value: val3, Meta: 57, UserMeta: 0})

	v := l.Get(util.KeyWithTs([]byte("key"), 0))
	require.True(t, v.Value == nil)

	v = l.Get(util.KeyWithTs([]byte("key1"), 0))
	require.True(t, v.Value != nil)
	require.EqualValues(t, "00042", string(v.Value))
	require.EqualValues(t, 55, v.Meta)

	v = l.Get(util.KeyWithTs([]byte("key2"), 0))
	require.True(t, v.Value == nil)

	v = l.Get(util.KeyWithTs([]byte("key3"), 0))
	require.True(t, v.Value != nil)
	require.EqualValues(t, "00062", string(v.Value))
	require.EqualValues(t, 57, v.Meta)

	l.Put(util.KeyWithTs([]byte("key3"), 1), kv.Value{Value: val4, Meta: 12, UserMeta: 0})
	v = l.Get(util.KeyWithTs([]byte("key3"), 1))
	require.True(t, v.Value != nil)
	require.EqualValues(t, "00072", string(v.Value))
	require.EqualValues(t, 12, v.Meta)

	l.Put(util.KeyWithTs([]byte("key4"), 1), kv.Value{Value: val5, Meta: 60, UserMeta: 0})
	v = l.Get(util.KeyWithTs([]byte("key4"), 1))
	require.NotNil(t, v.Value)
	require.EqualValues(t, val5, v.Value)
	require.EqualValues(t, 60, v.Meta)
}
