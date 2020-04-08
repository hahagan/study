// study for list
package link_list

import (
	"fmt"
	"reflect"
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

type LinkList struct {
	values []interface{}
	length int
}

func (l *LinkList) Init(list ...interface{}) interface{} {
	l.values = []interface{}{}
	for i := range list {
		l.values = append(l.values, i)
		fmt.Println("item: ", reflect.ValueOf(i))
	}
	l.length = len(l.values)
	fmt.Printf("lenght: %d\n", l.length)
	return l
}

func (l *LinkList) Destroy() {
	l.values = nil
	l.length = -1
}

func (l *LinkList) Clear() {
	l.values = []interface{}{}
	l.length = 0
}

func (l *LinkList) Length() int {
	return l.length
}

func (l *LinkList) Get(index int) interface{} {
	if index > l.length-1 {
		panic(fmt.Sprintf("Get LinkList overload with index %d\n", index))
	}
	return l.values[index]
}

func (l *LinkList) Set(index int, i interface{}) {
	if index > l.length-1 {
		panic(fmt.Sprintf("Set LinkList overload with index %d\n", index))
	}
	l.values[index] = i
}
