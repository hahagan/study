package link

import (
	"fmt"
)

// 某一链表元素存储
// next指向链表的下一个元素，即指针域
// value为具体的数据，即数据域
type StaticListItem struct {
	value interface{}
	next  int
}

/*
根据数据结构一书中的实现，其数组头部记录的是数组中空闲位置组成的数组头部
但是如此的实现方式无法获取链表头部，从而不能从头部开始进行遍历，获取指定位置的元素
因此增加了free和head，分别指向空闲数组头部与链表头部
capacity用于初始化分配空间大小，capacityBase可在数组满时按其值增加数组大小
*/
type StaticList struct {
	free         int
	length       int
	capacity     int
	head         int
	capacityBase int
	values       []StaticListItem
}

// 创建连续数组，并赋值给链表中数组存储遍历，时间复杂度O(n)
func (l *StaticList) Init(capacity int, capacityBase int) *StaticList {
	l.values = make([]StaticListItem, capacity)
	for i := 0; i < capacity-1; i++ {
		l.values[i].next = i + 1
	}
	l.values[capacity-1].next = -1

	l.length = 0
	l.capacity = capacity
	l.free = 0
	l.head = -1
	l.capacityBase = capacityBase
	return l
}

// 有点粗暴，将链表数组部分设为空，世界复杂度O(1)
func (l *StaticList) Destroy() {
	l.length = -1
	l.capacity = 0
	l.head = -1
	l.values = nil
	l.free = -1
}

// 将所有链表元素加入空闲数组，时间复杂度O(n)
func (l *StaticList) Clear() {
	capacity := l.capacity
	for i := 0; i < capacity-1; i++ {
		l.values[i].next = i + 1
	}
	l.values[capacity-1].next = -1
	l.length = 0
	l.free = 0
	l.head = -1
}

func (l *StaticList) Length() int {
	return l.length
}

// 从head指向的头部遍历index次，获取第index个值，时间复杂度为O(n)
func (l *StaticList) get(index int) (int, StaticListItem) {
	if index > l.length-1 {
		panic(fmt.Sprintf("Get LinkList overload with index %d\n", index))
	}

	cur := l.values[l.head]
	indexCur := l.head
	for i := 1; i <= index; i++ {
		indexCur = cur.next
		cur = l.values[cur.next]
	}

	return indexCur, cur
}

//调用获取链表第index元素，并返回元素数据域
func (l *StaticList) Get(index int) interface{} {
	_, cur := l.get(index)
	return cur.value
}

// 链表扩容，时间复杂度O(n)
func (l *StaticList) expand(capacity int) {
	values := make([]StaticListItem, capacity)
	for i := 0; i < l.length; i++ {
		values[i] = l.values[i]
	}
	for i := l.length; i < capacity-1; i++ {
		values[i].next = i + 1
	}
	values[capacity-1].next = -1
	l.values = values
	l.free = l.length
	l.capacity = capacity
}

// 申请一个空闲链表元素，剩余空间不足时自动扩容，时间复杂度为O(1)或O(n)
func (l *StaticList) mallocFree() (int, StaticListItem) {
	var free StaticListItem
	var freeIndex int
	if l.free != -1 {
		free = l.values[l.free]
		freeIndex = l.free
		l.free = free.next
	} else {
		l.expand(l.capacity + l.capacityBase)
		free = l.values[l.free]
		freeIndex = l.free
		l.free = free.next
	}

	return freeIndex, free
}

// 获取链表指定位置元素前驱，并将数据插入链表第index位置， 时间复杂度O(n)
func (l *StaticList) Insert(index int, i interface{}) {
	if index > l.length-1 {
		index = l.length - 1
	} else if index <= 0 {
		index = 0
	}

	freeIndex, _ := l.mallocFree()
	l.values[freeIndex].value = i
	if index <= 0 {
		l.values[freeIndex].next = l.head
		l.head = freeIndex
	} else {
		indexPrev, prev := l.get(index - 1)
		l.values[freeIndex].next = prev.next
		l.values[indexPrev].next = freeIndex
	}

	l.length += 1
}

// 释放指定数组位置的数据，使其加入空闲链表，时间复杂度O(1)
func (l *StaticList) freeItem(index int) {
	l.values[index].next = l.free
	l.free = index
}

// 删除链表第index个元素，时间复杂度O(n)
func (l *StaticList) Delete(index int) {
	if index > l.length-1 || index < 0 {
		panic(fmt.Sprintf("Delete LinkList overload with index %d, only within %d\n", index, l.length))
	}
	_, prev := l.get(index - 1)
	prev.next = l.values[prev.next].next
	l.freeItem(index)
	l.length -= 1
}

// 设置链表第index个元素值，时间复杂度O(n)
func (l *StaticList) Set(index int, i interface{}) {
	if index > l.length-1 || index < 0 {
		panic(fmt.Sprintf("Delete PointerList overload with index %d, only within %d\n", index, l.length))
	}
	targetIndex, _ := l.get(index)
	l.values[targetIndex].value = i
}
