//连续存储
package link

import (
	"fmt"
)

type ArrayLinkList struct {
	values       *[]interface{}
	length       int
	capacity     int
	capacityBase int
}

func (l *ArrayLinkList) Init(base int, list ...interface{}) interface{} {
	// fmt.Println(" ")
	length := len(list)
	var capacity int
	if length > base {
		capacity = int((length/base + 1) * base)
	} else {
		capacity = base
	}

	l.capacity = capacity
	l.length = length
	values := make([]interface{}, capacity)

	index := 0
	for i := range list {
		values[index] = i
		index++
	}
	l.values = &values
	return l
}

func (l *ArrayLinkList) Destroy() error {
	l.values = nil
	l.length = -1
	l.capacity = -1
	l.capacityBase = -1
	return nil
}

func (l *ArrayLinkList) Clear() error {
	l.length = 0
	return nil
}

func (l *ArrayLinkList) Length() int {
	return l.length
}

func (l *ArrayLinkList) Get(index int) interface{} {
	if index > l.length-1 || l.values == nil {
		panic(fmt.Sprintf("Get ArrayList overload with index %d\n", index))
	}
	tmp := *l.values
	return tmp[index]
}

func (l *ArrayLinkList) expand(capacity int) {
	values := make([]interface{}, capacity)
	for i := 0; i < l.length; i++ {
		values[i] = l.Get(i)
	}
	l.values = &values
	l.capacity = capacity
}

func (l *ArrayLinkList) Insert(index int, i interface{}) int {
	if l.length+1 == l.capacity {
		l.expand(l.capacity + l.capacityBase)
	}

	if index > l.length-1 {
		index = l.length - 1
	} else if index <= 0 {
		index = 0
	}

	if index < 0 {
		index = 0
	}

	tmp := *l.values
	for j := l.length - 1; j >= index; j-- {
		tmp[j+1] = tmp[j]
	}

	tmp[index] = i
	l.length += 1
	return l.length
}

func (l *ArrayLinkList) Delete(index int) int {
	if index > l.length-1 || index < 0 {
		panic(fmt.Sprintf("Delete LinkList overload with index %d, only within %d\n", index, l.length))
	}

	tmp := *l.values
	for j := index; j < l.length-1; j++ {
		tmp[j] = tmp[j+1]
	}
	l.length -= 1
	return l.length
}

func (l *ArrayLinkList) Set(index int, i interface{}) {
	if index > l.length-1 {
		panic(fmt.Sprintf("Set LinkList overload with index %d\n", index))
	}
	tmp := *l.values
	tmp[index] = i
}
