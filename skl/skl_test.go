package skl

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestSimpleSKL(t *testing.T) {
	// 初始化跳表
	sl := NewSkipList()
	rand.NewSource(time.Now().UnixNano())

	// 插入数据
	sl.Insert(1)
	sl.Insert(2)
	sl.Insert(3)
	sl.Insert(4)

	// 查找数据
	if node := sl.Search(2); node != nil {
		fmt.Println("Found:", node.value)
	} else {
		fmt.Println("Not Found")
	}

	// 删除数据
	sl.Delete(3)
	if node := sl.Search(3); node != nil {
		fmt.Println("Found:", node.value)
	} else {
		fmt.Println("Not Found")
	}
}
