package skl

import "math/rand"

const (
	maxHeight = 20
)

type Node struct {
	value int
	next  []*Node
}

type SkipList struct {
	head  *Node
	level int
}

func newNode(val int, level int) *Node {
	return &Node{
		value: val,
		next:  make([]*Node, level),
	}
}

func NewSkipList() *SkipList {
	return &SkipList{
		head:  newNode(0, maxHeight),
		level: 1,
	}
}

// Insert 向跳表插入一个值
func (sl *SkipList) Insert(value int) {
	update := make([]*Node, maxHeight)
	current := sl.head

	// 从最高层开始查找插入位置
	for i := sl.level - 1; i >= 0; i-- {
		for current.next[i] != nil && current.next[i].value < value {
			current = current.next[i]
		}
		update[i] = current
	}

	// 如果下层的节点值等于要插入的值，直接返回
	current = current.next[0]
	if current != nil && current.value == value {
		return
	}

	// 随机决定新节点的层数
	level := sl.RandomLevel()

	// 如果新节点的层数大于当前跳表的层数，更新跳表的层数
	if level > sl.level {
		for i := sl.level; i < level; i++ {
			update[i] = sl.head
		}
		sl.level = level
	}

	// 创建新节点
	newNode := newNode(value, level)

	// 更新新节点的指针
	for i := 0; i < level; i++ {
		newNode.next[i] = update[i].next[i]
		update[i].next[i] = newNode
	}
}

// RandomLevel 随机生成一个层数
func (sl *SkipList) RandomLevel() int {
	level := 1
	for rand.Float32() < 0.5 && level < maxHeight {
		level++
	}
	return level
}

// Search 查找跳表中的值
func (sl *SkipList) Search(value int) *Node {
	current := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for current.next[i] != nil && current.next[i].value < value {
			current = current.next[i]
		}
	}

	// 查找最底层的节点
	current = current.next[0]
	if current != nil && current.value == value {
		return current
	}
	return nil
}

// Delete 删除跳表中的值
func (sl *SkipList) Delete(value int) {
	update := make([]*Node, maxHeight)
	current := sl.head

	// 从最高层开始查找删除位置
	for i := sl.level - 1; i >= 0; i-- {
		for current.next[i] != nil && current.next[i].value < value {
			current = current.next[i]
		}
		update[i] = current
	}

	// 查找最底层的节点
	current = current.next[0]
	if current != nil && current.value == value {
		// 更新各层的指针
		for i := 0; i < sl.level; i++ {
			if update[i].next[i] != current {
				break
			}
			update[i].next[i] = current.next[i]
		}

		// 可能需要更新跳表的层数
		for sl.level > 1 && sl.head.next[sl.level-1] == nil {
			sl.level--
		}
	}
}
