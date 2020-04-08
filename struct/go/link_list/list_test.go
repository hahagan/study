package link_list

import (
	"testing"
)

func TestLinkList(t *testing.T) {
	l := new(LinkList)
	testInit(t, l, 0, 1, 2, 3)
	testLength(t, l, 4)
	testGet(t, l, 2)
	testSet(t, l, 3, 4)
	testClear(t, l)
	testDestroy(t, l)
}

func testInit(t *testing.T, l *LinkList, args ...int) {
	list := []interface{}{}
	for i := range args {
		list = append(list, interface{}(i))
	}
	l.Init(list...)
}

func testLength(t *testing.T, l *LinkList, length int) {
	length1 := l.Length()
	if length1 != length {
		t.Error("TestLength error")
	}

}

func testClear(t *testing.T, l *LinkList) {
	l.Clear()
	if l.values == nil || l.length != 0 {
		t.Error("TestClear error")
	}
}

func testDestroy(t *testing.T, l *LinkList) {
	l.Destroy()
	if l.values != nil || l.length != -1 {
		t.Error("TestDestroy error")
	}
}

func testGet(t *testing.T, l *LinkList, index int) {
	if index != l.Get(index) {
		t.Error("TestGet error")
	}
}

func testSet(t *testing.T, l *LinkList, index int, i int) {
	l.Set(index, i)
	if i != l.Get(index) {
		t.Errorf("TestSet error, set: %d, get:　%d\n", i, l.Get(index))
	}
}
