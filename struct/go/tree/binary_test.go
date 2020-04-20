package tree

import (
	"fmt"
	"testing"

	"github.com/hahagan/study/struct/go/list/link"
)

func TestBinaryTree(t *testing.T) {
	tmp := new(BinaryTree)
	tmp.Init()
	testBinaryTreeDepth(t, tmp, 0)
	testBinaryTreeLength(t, tmp, 0)
	times := 2
	for i := 0; i < times; i++ {
		testBinaryTreeInsert(t, tmp, i)
	}

	testBinaryTreeInsert(t, tmp, -1)

	testBinaryTreeLength(t, tmp, 3)
	testBinaryTreeDepth(t, tmp, times)

	fmt.Println("testBinaryTreePrevOrderVist")
	want := []int{0, -1, 1}
	testBinaryTreePrevOrderVist(t, tmp, want)

	fmt.Println("testBinaryTreeInOrderVist")
	want = []int{-1, 0, 1}
	testBinaryTreeInOrderVist(t, tmp, want)

	fmt.Println("testBinaryTreePostOrderVist")
	want = []int{-1, 1, 0}
	testBinaryTreePostOrderVist(t, tmp, want)
}

func testBinaryTreeInsert(t *testing.T, tmp *BinaryTree, v int) {
	lengthOld := tmp.Length()
	tmp.Insert(v, v)
	lengthNew := tmp.Length()
	if lengthOld != lengthNew-1 {
		t.Errorf("testBinaryTreeInsert error, before insert length :%d, after insert length: %d", lengthOld, lengthNew)
	}

	err := tmp.Insert(v, v)
	if err == nil {
		t.Errorf(
			"testBinaryTreeInsert error, Repeated insertion without error, index: %d", v)
	}
}

func testBinaryTreeLength(t *testing.T, tmp *BinaryTree, v int) {
	if tmp.Length() != v {
		t.Errorf("testBinaryTreeLength error, want :%d, get: %d", v, tmp.Length())
	}
}

func testBinaryTreeDepth(t *testing.T, tmp *BinaryTree, v int) {
	d := tmp.Depth()
	if d != v {
		t.Errorf("testBinaryTreeDepth error, want :%d, get: %d, root: %+v", v, d, tmp.root)
	}
}

func testBinaryTreePrevOrderVist(t *testing.T, tmp *BinaryTree, want []int) {
	r := new(link.ArrayLinkList).Init(10)
	tmp.PrevOrderVist(resultSave(r))
	for i := 0; i < r.Length(); i++ {
		if r.Get(i).(int) != want[i] {
			t.Errorf("testBinaryTreePrevOrderVist, want %d: %d, but get %d", i, want[i], r.Get(i).(int))
		}
	}
}

func testBinaryTreeInOrderVist(t *testing.T, tmp *BinaryTree, want []int) {
	r := new(link.ArrayLinkList).Init(10)
	tmp.InOrderVist(resultSave(r))
	for i := 0; i < r.Length(); i++ {
		if r.Get(i).(int) != want[i] {
			t.Errorf(
				"testBinaryTreePrevOrderVist, want %d: %d, but get %d",
				i, want[i], r.Get(i).(int))
		}
	}
}

func testBinaryTreePostOrderVist(t *testing.T, tmp *BinaryTree, want []int) {
	r := new(link.ArrayLinkList).Init(10)
	tmp.PostOrderVist(resultSave(r))
	for i := 0; i < r.Length(); i++ {
		if r.Get(i).(int) != want[i] || r.Length() != len((want)) {
			t.Errorf(
				"%s, want %d: %d, get %d, result length: %d, want length: %d",
				"testBinaryTreePrevOrderVist",
				i, want[i], r.Get(i).(int), r.Length(), len(want))
		}
	}
}
