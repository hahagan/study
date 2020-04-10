// 连续存储
package link

import (
	"fmt"
	// "reflect"
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

//Implement linklist with slice
//Unsafe, because of i can't controll slice in memmory
//Just for fun
type SliceLinkList struct {
	values []interface{}
	length int
}

func (l *SliceLinkList) Init(list ...interface{}) interface{} {
	l.values = []interface{}{}
	for i := range list {
		l.values = append(l.values, i)
		// fmt.Println("item: ", reflect.ValueOf(i))
	}
	l.length = len(l.values)
	return l
}

func (l *SliceLinkList) Destroy() {
	l.values = nil
	l.length = -1
}

func (l *SliceLinkList) Clear() {
	l.values = []interface{}{}
	l.length = 0
}

func (l *SliceLinkList) Length() int {
	return l.length
}

func (l *SliceLinkList) Get(index int) interface{} {
	if index > l.length-1 {
		panic(fmt.Sprintf("Get LinkList overload with index %d\n", index))
	}
	return l.values[index]
}

func (l *SliceLinkList) Insert(index int, i interface{}) int {
	if index > l.length-1 {
		index = l.length - 1
	} else if index <= 0 {
		index = 0
	}
	v := []interface{}{}
	for j := 0; j < l.length; j++ {
		if j == index {
			v = append(v, i)
		}
		v = append(v, l.values[j])
	}

	if l.length == 0 {
		v = append(v, i)
	}

	l.values = v
	// for j := 0; j < l.length+1; j++ {
	// 	fmt.Println("|---- item: ", v[j])
	// }
	l.length += 1
	return l.length
}

func (l *SliceLinkList) Delete(index int) int {
	if index > l.length-1 || index < 0 {
		panic(fmt.Sprintf("Delete LinkList overload with index %d, only within %d\n", index, l.length))
	}

	if index == l.length-1 {
		l.length -= 1
	} else {
		l.values = append([]interface{}{}, l.values[:index], l.values[index+1:])
		l.length -= 1
	}

	return l.length
}

func (l *SliceLinkList) Set(index int, i interface{}) {
	if index > l.length-1 {
		panic(fmt.Sprintf("Set LinkList overload with index %d\n", index))
	}
	l.values[index] = i
}
