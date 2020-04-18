package tree

import (
	"fmt"
	"testing"

	"github.com/hahagan/study/struct/go/list/link"
)

func resultSave(r *link.ArrayLinkList) func(interface{}) error {

	return func(v interface{}) error {
		_, err := fmt.Println("|--------------- ", v)
		r.Insert(r.Length()+1, v)
		return err
	}
}

func TestBianaryTree(t *testing.T) {
	tmp := new(BinaryTree)
	tmp.Init()
	testBianaryTreeDepth(t, tmp, 0)
	testBianaryTreeLength(t, tmp, 0)
	times := 2
	for i := 0; i < times; i++ {
		testBianaryTreeInsert(t, tmp, i)
	}

	testBianaryTreeInsert(t, tmp, -1)

	testBianaryTreeLength(t, tmp, 3)
	testBianaryTreeDepth(t, tmp, times)
	want := []int{0, -1, 1}
	testPrevOrderVist(t, tmp, want)
	want = []int{-1, 0, 1}
	testInOrderVist(t, tmp, want)
}

func testBianaryTreeInsert(t *testing.T, tmp *BinaryTree, v int) {
	lengthOld := tmp.Length()
	tmp.insert(v, v)
	lengthNew := tmp.Length()
	if lengthOld != lengthNew-1 {
		t.Errorf("testBianaryTreeInsert error, before insert length :%d, after insert length: %d", lengthOld, lengthNew)
	}
}

func testBianaryTreeLength(t *testing.T, tmp *BinaryTree, v int) {
	if tmp.Length() != v {
		t.Errorf("testBianaryTreeLength error, want :%d, get: %d", v, tmp.Length())
	}
}

func testBianaryTreeDepth(t *testing.T, tmp *BinaryTree, v int) {
	d := tmp.Depth()
	if d != v {
		t.Errorf("testBianaryTreeDepth error, want :%d, get: %d", v, d)
	}
}

func testPrevOrderVist(t *testing.T, tmp *BinaryTree, want []int) {
	r := new(link.ArrayLinkList).Init(10)
	tmp.PrevOrderVist(resultSave(r))
	for i := 0; i < r.Length(); i++ {
		if r.Get(i).(int) != want[i] {
			t.Errorf("testPrevOrderVist, want %d: %d, but get %d", i, want[i], r.Get(i).(int))
		}
	}
}

func testInOrderVist(t *testing.T, tmp *BinaryTree, want []int) {
	r := new(link.ArrayLinkList).Init(10)
	tmp.root.InOrderVist(resultSave(r))
	for i := 0; i < r.Length(); i++ {
		if r.Get(i).(int) != want[i] {
			t.Errorf("testPrevOrderVist, want %d: %d, but get %d", i, want[i], r.Get(i).(int))
		}
	}
}