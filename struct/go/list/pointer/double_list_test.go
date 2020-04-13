package pointer

import (
	"fmt"
	"reflect"
	"testing"
)

func TestDoubleList(t *testing.T) {
	fmt.Println("test DoubleList: ")
	l := new(DoubleList)
	testDoubleListInit(t, l)
	for i := 3; i >= 0; i-- {
		testDoubleListInsert(t, l, 0, i)
	}

	testDoubleListLength(t, l, 4)
	testDoubleListGet(t, l, 2)
	testDoubleListSet(t, l, 3, 4)
	testDoubleListInsert(t, l, 3, 3)
	testDoubleListInsert(t, l, 6, 5)
	testDoubleListDelete(t, l, 4)
	testDoubleListClear(t, l)
	testDoubleListInsert(t, l, 0, 0)
	testDoubleListInsert(t, l, -3, 0)
	testDoubleListDelete(t, l, 0)
	testDoubleListDestroy(t, l)
	l = new(DoubleList)
	testDoubleListInit(t, l)
	for i := 20; i >= 0; i-- {
		testDoubleListInsert(t, l, 0, i)
	}
	for i := 20; i >= 0; i-- {
		testDoubleListDelete(t, l, 0)
	}

}

func testDoubleListInit(t *testing.T, l *DoubleList) {
	l.Init()
	if l.length != 0 || l.Next != nil || l.value != nil || l.Prev != nil {
		t.Error("Test DoubleList Init() error")
	}
}

func testDoubleListDestroy(t *testing.T, l *DoubleList) {
	l.Destroy()
	if l.length != -1 || l.Next != nil || l.value != nil || l.Prev != nil {
		t.Error("Test DoubleList Destroy() error", l)
	}
}

func testDoubleListClear(t *testing.T, l *DoubleList) {
	l.Clear()
	if l.length != 0 || l.Next != nil || l.value != nil {
		t.Error("Test DoubleList Clear() error")
	}
}

func testDoubleListLength(t *testing.T, l *DoubleList, length int) {
	length1 := l.Length()
	if length1 != length {
		t.Error("Test DoubleList Length() error")
	}
}

func testDoubleListGet(t *testing.T, l *DoubleList, index int) {
	if v := l.Get(index); index != v {
		t.Errorf("Test DoubleList Get %d error, v: %d", index, v)
		cur := l
		for cur.Next != nil {
			cur = cur.Next
			v := cur.value
			t.Error("|---- item: ", reflect.ValueOf(v))
		}
	}
}

func testDoubleListInsert(t *testing.T, l *DoubleList, index int, i int) {
	length := l.Length()
	l.Insert(index, i)
	if index > l.length-1 {
		index = l.length - 2
	}
	if index < 0 {
		index = 0
	}

	if item := l.Get(index); i != item || l.Length() != length+1 {
		t.Errorf("Test DoubleList insert error, insert: %d, get item %d:　%d\n", i, index, item)
		cur := l
		for cur.Next != nil {
			cur = cur.Next
			v := cur.value
			t.Error("|---- item: ", reflect.ValueOf(v))
		}
	}
}

func testDoubleListDelete(t *testing.T, l *DoubleList, index int) {
	oldLength := l.length
	newLength := 0
	l.Delete(index)
	for cur := l; cur.Next != nil; cur = cur.Next {
		newLength++
	}
	if newLength != oldLength-1 {
		t.Error("Test DoubleList Delete error")
	}
}

func testDoubleListSet(t *testing.T, l *DoubleList, index int, i int) {
	l.Set(index, i)
	if v := l.Get(index); i != v {
		t.Errorf("Test DoubleList Set error, set: %d, get:　%d\n", i, v)
	}
}
