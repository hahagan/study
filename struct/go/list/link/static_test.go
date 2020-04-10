package link

import (
	"fmt"
	"reflect"
	"testing"
)

func TestStaticList(t *testing.T) {
	fmt.Println("test StaticList: ")
	l := new(StaticList)
	testStaticListInit(t, l, 4, 4)
	for i := 3; i >= 0; i-- {
		testStaticListInsert(t, l, 0, i)
	}

	testStaticListLength(t, l, 4)
	testStaticListGet(t, l, 2)
	testStaticListSet(t, l, 3, 4)
	testStaticListInsert(t, l, 3, 4)
	testStaticListInsert(t, l, 6, 5)
	testStaticListDelete(t, l, 4)
	testStaticListClear(t, l)
	testStaticListInsert(t, l, 0, 0)
	testStaticListInsert(t, l, -3, 0)
	testStaticListDelete(t, l, 0)
	testStaticListDestroy(t, l)

}

func testStaticListInit(t *testing.T, l *StaticList, capacity int, capacityBase int) {
	l.Init(capacity, capacityBase)
	if l.length != 0 || l.head != -1 || l.free != 0 || cap(l.values) != capacity {
		t.Error("Test StaticList Init() error", l.length, l.head, l.free, cap(l.values), capacity)
	}
}

func testStaticListDestroy(t *testing.T, l *StaticList) {
	l.Destroy()
	if l.length != -1 || l.capacity != 0 || l.values != nil || l.head != -1 || l.free != -1 {
		t.Error("Test StaticList Destroy() error", l.length, l.capacity, l.head, l.free)
	}
}

func testStaticListClear(t *testing.T, l *StaticList) {
	l.Clear()
	if l.length != 0 || l.free != 0 || l.head != -1 {
		t.Error("Test StaticList Clear() error")
	}
}

func testStaticListLength(t *testing.T, l *StaticList, length int) {
	length1 := l.Length()
	if length1 != length {
		t.Error("Test StaticList Length() error")
	}
}

func testStaticListGet(t *testing.T, l *StaticList, index int) {
	if index != l.Get(index) {
		t.Error("Test StaticList Get error")
	}
}

func testStaticListInsert(t *testing.T, l *StaticList, index int, i int) {
	length := l.Length()
	l.Insert(index, i)
	if index > l.length-1 {
		index = l.length - 2
	}
	if index < 0 {
		index = 0
	}

	if i != l.Get(index) || l.Length() != length+1 {
		t.Errorf("Test StaticList insert error, insert: %d, get item %d:　%d\n", i, index, l.Get(index))
		for j := 0; j < l.length; j++ {
			t.Error("|---- item: ", reflect.ValueOf(l.Get(j)))
		}
	}
}

func testStaticListDelete(t *testing.T, l *StaticList, index int) {
	l.Delete(index)
}

func testStaticListSet(t *testing.T, l *StaticList, index int, i int) {
	l.Set(index, i)
	if i != l.Get(index) {
		t.Errorf("Test StaticList Set error, set: %d, get:　%d\n", i, l.Get(index))
	}
}
