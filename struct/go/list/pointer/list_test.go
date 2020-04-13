package pointer

import (
	"fmt"
	"reflect"
	"testing"
)

func TestPointerList(t *testing.T) {
	fmt.Println("test ArrayList: ")
	l := new(PointerList)
	testPointerListInit(t, l)
	for i := 3; i >= 0; i-- {
		testPointerListInsert(t, l, 0, i)
	}

	testPointerListLength(t, l, 4)
	testPointerListGet(t, l, 2)
	testPointerListSet(t, l, 3, 4)
	testPointerListInsert(t, l, 3, 3)
	testPointerListInsert(t, l, 6, 5)
	testPointerListDelete(t, l, 4)
	testPointerListClear(t, l)
	testPointerListInsert(t, l, 0, 0)
	testPointerListInsert(t, l, -3, 0)
	testPointerListDelete(t, l, 0)
	testPointerListDestroy(t, l)

}

func testPointerListInit(t *testing.T, l *PointerList) {
	l.Init()
	if l.length != 0 || l.next != nil || l.value != nil {
		t.Error("Test PointerList Init() error")
	}
}

func testPointerListDestroy(t *testing.T, l *PointerList) {
	l.Destroy()
	if l.length != -1 || l.next != nil || l.value != nil {
		t.Error("Test PointerList Destroy() error", l)
	}
}

func testPointerListClear(t *testing.T, l *PointerList) {
	l.Clear()
	if l.length != 0 || l.next != nil || l.value != nil {
		t.Error("Test PointerList Clear() error")
	}
}

func testPointerListLength(t *testing.T, l *PointerList, length int) {
	length1 := l.Length()
	if length1 != length {
		t.Error("Test PointerList Length() error")
	}
}

func testPointerListGet(t *testing.T, l *PointerList, index int) {
	if index != l.Get(index) {
		t.Error("Test PointerList Get error")
	}
}

func testPointerListInsert(t *testing.T, l *PointerList, index int, i int) {
	length := l.Length()
	l.Insert(index, i)
	if index > l.length-1 {
		index = l.length - 2
	}
	if index < 0 {
		index = 0
	}
	if i != l.Get(index) || l.Length() != length+1 {
		t.Errorf("Test PointerList insert error, insert: %d, get item %d:　%d\n", i, index, l.Get(index))
		for j := 0; j < l.length; j++ {
			t.Error("|---- item: ", reflect.ValueOf(l.Get(j)))
		}
	}
}

func testPointerListDelete(t *testing.T, l *PointerList, index int) {
	l.Delete(index)
}

func testPointerListSet(t *testing.T, l *PointerList, index int, i int) {
	l.Set(index, i)
	if i != l.Get(index) {
		t.Errorf("Test PointerList Set error, set: %d, get:　%d\n", i, l.Get(index))
	}
}
