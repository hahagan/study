// array_test
package link

import (
	"fmt"
	"reflect"
	"testing"
)

func TestArrayList(t *testing.T) {
	fmt.Println("test ArrayList: ")
	l := new(ArrayLinkList)
	testArrayInit(t, l, 2, 0, 1, 2, 3)
	testArrayLength(t, l, 4)
	testArrayGet(t, l, 2)
	testArraySet(t, l, 3, 4)
	testArrayInsert(t, l, 3, 3)
	testArrayInsert(t, l, 6, 5)
	testArrayDelete(t, l, 4)
	testArrayClear(t, l)
	testArrayInsert(t, l, 0, 0)
	testArrayInsert(t, l, -3, 0)
	testArrayDelete(t, l, 0)
	testArrayDestroy(t, l)
	testArrayInit(t, l, 10, 0, 1, 2, 3)
	for i := 0; i < 12; i++ {
		testArrayInsert(t, l, i, i)
	}
	testArray_expand(t, l)

}

func testArray_expand(t *testing.T, l *ArrayLinkList) {
	want := l.capacity + l.capacityBase
	l.expand(want)
	tmp := *l.values
	if l.capacity != cap(tmp) || l.capacity != want {
		t.Errorf("Test Array expand error, want: %d, get, %d", want, l.capacity)
	}
}

func testArrayInit(t *testing.T, l *ArrayLinkList, base int, args ...int) {
	list := []interface{}{}
	for i := range args {
		list = append(list, interface{}(i))
	}
	l.Init(base, list...)
	length := len(list)

	var capacity int
	if length > base {
		capacity = int((length/base + 1) * base)
	} else {
		capacity = base
	}
	if l.length != len(args) || l.capacity != capacity {
		t.Error("Test Array Init() error")
	}
}

func testArrayDestroy(t *testing.T, l *ArrayLinkList) {
	l.Destroy()
	if l.length != -1 || l.capacity != -1 || l.capacityBase != -1 || l.values != nil {
		t.Error("Test Array Destroy() error")
	}
}

func testArrayClear(t *testing.T, l *ArrayLinkList) {
	l.Clear()
	if l.length != 0 || l.values == nil {
		t.Error("Test Array Clear() error")
	}
}

func testArrayLength(t *testing.T, l *ArrayLinkList, length int) {
	length1 := l.Length()
	if length1 != length {
		t.Error("Test Array Length() error")
	}
}

func testArrayGet(t *testing.T, l *ArrayLinkList, index int) {
	if index != l.Get(index) {
		t.Error("Test Array Get error")
	}
}

func testArrayInsert(t *testing.T, l *ArrayLinkList, index int, i int) {
	length := l.Length()
	l.Insert(index, i)
	if index > l.length-1 {
		index = l.length - 1
	}
	if index < 0 {
		index = 0
	}
	if i != l.Get(index) || l.Length() != length+1 {
		t.Errorf("Testinsert error, insert: %d, get %d:　%d\n", i, index, l.Get(index))
		tmp := *l.values
		for j := 0; j < l.length; j++ {
			t.Error("|---- item: ", reflect.ValueOf(tmp[j]))
		}
	}
}

func testArrayDelete(t *testing.T, l *ArrayLinkList, index int) {
	l.Delete(index)
}

func testArraySet(t *testing.T, l *ArrayLinkList, index int, i int) {
	l.Set(index, i)
	if i != l.Get(index) {
		t.Errorf("TestSet error, set: %d, get:　%d\n", i, l.Get(index))
	}
}
