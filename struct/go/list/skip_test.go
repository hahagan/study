package list

import (
	"fmt"
	"testing"
)

func TestSkipList(t *testing.T) {
	fmt.Println("TestSkipList:")
	l := testSkipListInit(t, 20)
	for i := 0; i <= 1000; i++ {
		testSkipListInsert(t, l, i, i)
	}
	testSkipListInsert(t, l, 0, 0)
	testSkipListInsert(t, l, 1, 1)
	for i := 0; i <= 500; i++ {
		testSkipListDelete(t, l, i)
	}

	for i := 1000; i >= 501; i-- {
		testSkipListDelete(t, l, i)
	}

	l = testSkipListInit(t, -1)

}

func testSkipListInit(t *testing.T, level int) *SkipList {
	l := new(SkipList).Init(level)
	if level <= 0 {
		level = 32
	}
	if level_cap := cap(l.head.next); level_cap != level {
		t.Errorf("Init Error, level: %d, get: %d", level, level_cap)
	}
	return l
}

func testSkipListInsert(t *testing.T, l *SkipList, order int, v interface{}) {
	l.Insert(order, v)
	want, err := l.Find(order)
	if v != want || err != nil {
		t.Errorf("Find or Insert error, order: %d, want %d, error: %v", order, v, err)
		t.Error("|---- get: ", want)
	}
}

func testSkipListFind(t *testing.T, l *SkipList, order int, want int) {
	v, err := l.Find(order)
	if v != want || err != nil {
		t.Errorf("Find errot, want %d, get %d", v, want)
	}
}

func testSkipListDelete(t *testing.T, l *SkipList, order int) {
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		fmt.Printf("Delete panic, delete %d\n%v", order, err)
	// 	}

	// }()
	l.Delete(order)
	_, err := l.Find(order)
	if err == nil {
		t.Errorf("Delete error, delete %d", order)
	}
}
