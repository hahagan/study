// 链式指针存储方式
package pointer

import (
	"fmt"
)

type List interface {
	Init(interface{}) interface{}
	Destroy() error
	Clear() error
	Length() int
	Get(int) interface{}
	Insert(int, interface{}) error
	Delete(int) error
	set(int, interface{})
}

type PointerList struct {
	value  interface{}
	next   *PointerList
	length int
}

func (l *PointerList) Init() *PointerList {
	l.value = nil
	l.next = nil
	l.length = 0
	return l
}

func (l *PointerList) Destroy() {
	l.length = -1
	l.next = nil
	l.value = nil
}

func (l *PointerList) Clear() {
	l.length = 0
	l.next = nil
	l.value = nil
}

func (l *PointerList) Length() int {
	return l.length
}

func (l *PointerList) Get(index int) interface{} {
	if index > l.length-1 {
		panic(fmt.Sprintf("Get LinkList overload with index %d\n", index))
	}

	cur := l.next
	for i := 0; i < index && cur != nil; i++ {
		cur = cur.next
	}

	if cur == nil {
		panic(fmt.Sprintf("Get LinkList overload with index %d\n", index))
	}

	return cur.value
}

func (l *PointerList) Insert(index int, i interface{}) {
	if index > l.length-1 {
		index = l.length - 1
	} else if index <= 0 {
		index = 0
	}

	if index < 0 {
		index = 0
	}

	cur := l
	for i := 0; i < index && cur != nil; i++ {
		cur = cur.next
	}
	item := new(PointerList).Init()

	item.value = i
	if cur.next != nil {
		item.next = l.next
		item.length = l.next.length + 1
	}

	cur.next = item
	cur.length += 1
	if cur != l {
		l.length += 1
	}
}

func (l *PointerList) Delete(index int) {
	if index > l.length-1 || index < 0 {
		panic(fmt.Sprintf("Delete LinkList overload with index %d, only within %d\n", index, l.length))
	}
	cur := l
	for i := 0; i < index && cur != nil; i++ {
		cur = cur.next
	}
	if cur.next != nil {
		next := cur.next.next
		cur.next = next
		cur.length -= 1
	}

	if l != cur {
		l.length -= 1
	}
}

func (l *PointerList) Set(index int, i interface{}) {
	if index > l.length-1 {
		panic(fmt.Sprintf("Get LinkList overload with index %d\n", index))
	}

	cur := l.next
	for i := 0; i < index && cur != nil; i++ {
		cur = cur.next
	}

	if cur == nil {
		panic(fmt.Sprintf("Get LinkList overload with index %d\n", index))
	}

	cur.value = i
}
