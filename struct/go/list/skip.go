package list

import (
	"fmt"
	"math/rand"
	"time"
)

// 该部分跳表代码实现基本是看着简书上的一位叫王大锤的作者提供的文档和代码进行理解并重写（基本等于抄）的，所以附上其原作者链接
// 作者：王大锤520
// 链接：https://www.jianshu.com/p/400d24e9daa0
// 来源：简书
// 著作权归作者所有。商业转载请联系作者获得授权，非商业转载请注明出处。
type SkipListNode struct {
	value interface{}
	order int
	next  []*SkipListNode
}

// 该部分跳表代码实现基本是看着简书上的一位叫王大锤的作者提供的文档和代码进行理解并重写（基本等于抄）的，所以附上其原作者链接
// 作者：王大锤520
// 链接：https://www.jianshu.com/p/400d24e9daa0
// 来源：简书
// 著作权归作者所有。商业转载请联系作者获得授权，非商业转载请注明出处。
type SkipList struct {
	head   *SkipListNode
	tail   *SkipListNode
	length int
	level  int
}

func (l *SkipList) Init(level int) *SkipList {
	l.length = 0
	if level <= 0 {
		level = 32
	}
	l.level = level
	l.tail = &SkipListNode{}
	l.head = &SkipListNode{
		next: make([]*SkipListNode, level),
	}
	for i := range l.head.next {
		l.head.next[i] = l.tail
	}
	rand.Seed(time.Now().Unix())

	return l
}

func (l *SkipList) randomLevel() int {
	level := 1
	for level < l.level && rand.Intn(2) != 1 {
		level++
	}
	return level
}

func (l *SkipList) Insert(order int, v interface{}) {
	level := l.randomLevel()
	index := make([]*SkipListNode, level, level)

	prev := l.head
	for i := level - 1; i >= 0; i-- {
		for {
			cur := prev.next[i]
			if cur == l.tail || cur.order > order {
				// 找到位置时，将前驱第i级索引设为前驱
				index[i] = prev
				break
			} else if cur.order == order {
				return
			} else {
				// 未找到位置，节点指针在第i级后移
				prev = cur
			}
		}

	}

	nodes := &SkipListNode{
		value: v,
		order: order,
		next:  make([]*SkipListNode, level, level),
	}

	for i, node := range index {
		node.next[i], nodes.next[i] = nodes, node.next[i]
	}
	l.length++

}

func (l *SkipList) find(order int) *SkipListNode {
	prev := l.head
	for i := len(prev.next) - 1; i >= 0; i-- {
		for {
			// fmt.Println(prev, i, order)
			cur := prev.next[i]
			if cur == l.tail || cur.order > order {
				break
			} else if cur.order == order {
				return cur
			} else {
				prev = cur
			}
		}
	}

	return nil
}

func (l *SkipList) Find(order int) (interface{}, error) {
	w := l.find(order)
	if w != nil {
		return w.value, nil
	}
	return nil, fmt.Errorf("Can't find!")
}

func (l *SkipList) Delete(order int) {

	prev := l.head
	index := make([]*SkipListNode, l.level, l.level)
	var taget *SkipListNode

	for i := len(prev.next) - 1; i >= 0; i-- {
		for {
			cur := prev.next[i]
			if cur == l.tail || cur.order > order {
				break
			} else if cur.order == order {
				index[i] = prev
				taget = cur
				break
			} else {
				prev = cur
			}
		}

	}

	for i, _ := range index {
		if index[i] != nil {
			index[i].next[i] = taget.next[i]

		}
	}

	if taget != nil {
		l.length -= 1
	}

}
