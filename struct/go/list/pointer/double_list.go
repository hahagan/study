package pointer

import (
	"fmt"
)

type DoubleList struct {
	value            interface{}
	Next, Prev, head *DoubleList
	length           int
}

func (l *DoubleList) Init() *DoubleList {
	l.value = nil
	l.head = l
	l.Next = nil
	l.Prev = nil
	l.length = 0
	return l
}

func (l *DoubleList) Destroy() {
	l.value = nil
	l.Next = nil
	l.Prev = nil
	l.length = -1
}

func (l *DoubleList) Clear() {
	l.value = nil
	l.Next = nil
	l.Prev = nil
	l.length = 0
}

func (l *DoubleList) Length() int {
	return l.length
}

func (l *DoubleList) Get(index int) interface{} {
	if index >= l.length {
		panic(fmt.Errorf("Out of range, index: %d, length: %d", index, l.length))
	}
	cur := l
	i := 0
	for i <= index && cur.Next != nil {
		cur = cur.Next
		i++
	}
	if i-1 != index {
		panic(fmt.Errorf("Out of range, index: %d, filnal: %d, %d", index, i-1, cur.value))
	}
	return cur.value
}

func (l *DoubleList) Insert(index int, i interface{}) error {
	if index > l.length-1 {
		index = l.length - 1
	} else if index <= 0 {
		index = 0
	}

	if index < 0 {
		index = 0
	}

	cur := l
	j := 0
	for j < index && cur.Next != nil {
		cur = cur.Next
		j++
	}

	next := cur.Next
	length := 0
	if next != nil {
		length = next.length + 1
	}
	item := DoubleList{
		value:  i,
		Next:   next,
		Prev:   cur,
		length: length,
		head:   l.head,
	}
	if next != nil {
		next.Prev = &item
	}
	cur.Next = &item
	cur.length += 1
	if cur != l.head {
		l.head.length += 1
	}
	return nil
}

func (l *DoubleList) Delete(index int) {
	if index >= l.length {
		panic(fmt.Errorf("Out of range, index: %d, length: %d", index, l.length))
	}

	j := 0
	cur := l
	for j <= index && cur.Next != nil {
		cur = cur.Next
		j++
	}

	if j-1 != index {
		panic(fmt.Errorf("Out of range, index: %d, filnal: %d", index, j-1))
	}

	prev := cur.Prev
	next := cur.Next
	prev.Next = next
	prev.length -= 1
	if next != nil {
		next.Prev = cur.Prev
		next.length = cur.length - 1
	}

	if l.head != prev {
		l.head.length -= 1
	}

}

func (l *DoubleList) Set(index int, item interface{}) error {
	if index >= l.length {
		return fmt.Errorf("Out of range, index: %d, length: %d", index, l.length)
	}

	i := 0
	cur := l
	for i <= index && cur.Next != nil {
		cur = cur.Next
		i++
	}

	if i-1 != index {
		return fmt.Errorf("Out of range, index: %d, filnal: %d", index, i)
	}

	cur.value = item
	return nil
}
