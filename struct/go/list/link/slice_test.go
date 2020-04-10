// 连续存储
package link

import (
	"fmt"
	"reflect"
	"testing"
)

func TestLinkList(t *testing.T) {
	l := new(SliceLinkList)
	fmt.Println("test LinkList: ")
	testInit(t, l, 0, 1, 2, 3)
	testLength(t, l, 4)
	testGet(t, l, 2)
	testSet(t, l, 3, 4)
	testInsert(t, l, 3, 3)
	testInsert(t, l, 6, 5)
	testDelete(t, l, 4)
	testClear(t, l)
	testInsert(t, l, 0, 0)
	testInsert(t, l, -3, 0)
	testDelete(t, l, 0)
	testDestroy(t, l)
}

func testInit(t *testing.T, l *SliceLinkList, args ...int) {
	list := []interface{}{}
	for i := range args {
		list = append(list, interface{}(i))
	}
	l.Init(list...)
}

func testLength(t *testing.T, l *SliceLinkList, length int) {
	length1 := l.Length()
	if length1 != length {
		t.Error("TestLength error")
	}

}

func testClear(t *testing.T, l *SliceLinkList) {
	l.Clear()
	if l.values == nil || l.length != 0 {
		t.Error("TestClear error")
	}
}

func testDestroy(t *testing.T, l *SliceLinkList) {
	l.Destroy()
	if l.values != nil || l.length != -1 {
		t.Error("TestDestroy error")
	}
}

func testGet(t *testing.T, l *SliceLinkList, index int) {
	if index != l.Get(index) {
		t.Error("TestGet error")
	}
}

func testInsert(t *testing.T, l *SliceLinkList, index int, i int) {
	length := l.Length()
	l.Insert(index, i)
	if index > l.length-1 {
		index = l.length - 2
	}
	if index < 0 {
		index = 0
	}
	if i != l.Get(index) || l.Length() != length+1 {
		t.Errorf("Testinsert error, insert: %d, get %d:　%d\n", i, index, l.Get(index))
		for j := 0; j < l.length; j++ {
			t.Error("|---- item: ", reflect.ValueOf(l.values[j]))
		}
	}
}

func testDelete(t *testing.T, l *SliceLinkList, index int) {
	l.Delete(index)
}

func testSet(t *testing.T, l *SliceLinkList, index int, i int) {
	l.Set(index, i)
	if i != l.Get(index) {
		t.Errorf("TestSet error, set: %d, get:　%d\n", i, l.Get(index))
	}
}
