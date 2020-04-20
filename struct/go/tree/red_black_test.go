package tree

import (
	"fmt"
	"testing"

	"math/rand"

	"github.com/hahagan/study/struct/go/list/link"
)

func resultSave(r *link.ArrayLinkList) func(interface{}) error {

	return func(v interface{}) error {
		// fmt.Println("|--------------- ", v)
		r.Insert(r.Length()+1, v)
		return nil
	}
}

func TestRedBlackTree(t *testing.T) {
	tmp := new(RedBlackTree)
	tmp.Init()
	times := 2

	testRedBlackTreeDepth(t, tmp, 0)
	testRedBlackTreeLength(t, tmp, 0)

	for i := 0; i < times; i++ {
		testRedBlackTreeInsert(t, tmp, i)
	}

	testRedBlackTreeInsert(t, tmp, -1)

	testRedBlackTreeLength(t, tmp, 3)
	testRedBlackTreeDepth(t, tmp, times)

	fmt.Println("testPrevOrderVist")
	want := []int{0, -1, 1}
	testPrevOrderVist(t, tmp, want)

	fmt.Println("testInOrderVist")
	want = []int{-1, 0, 1}
	testInOrderVist(t, tmp, want)

	fmt.Println("testPostOrderVist")
	want = []int{-1, 1, 0}
	testPostOrderVist(t, tmp, want)

	// 测试删除操作
	for i := 4; i < 1000; i++ {
		testRedBlackTreeInsert(t, tmp, rand.Intn(10000))
	}

	for i := 4; i < 1000; i++ {
		testRedBlackTreeDelete(t, tmp, rand.Intn(10000), want)
	}

	arr := []int{10, 7, 8, 15, 5, 6, 11, 13, 12}
	tmp1 := new(RedBlackTree)
	tmp1.Init()
	tmp1.locateNode(0)
	for i := 0; i < len(arr); i++ {
		tmp1.Insert(arr[i], arr[i])
	}

	tmp1.locateNode(10)
	tmp1.Delete(10)
	tmp1.Delete(21)
}

func testRedBlackTreeInsert(t *testing.T, tmp *RedBlackTree, v int) {
	lengthOld := tmp.Length()
	err1 := tmp.Insert(v, v)
	lengthNew := tmp.Length()
	if lengthOld != lengthNew-1 && err1 == nil {
		t.Errorf("testRedBlackTreeInsert error, before insert length :%d, after insert length: %d", lengthOld, lengthNew)
	}

	// err := tmp.Insert(v, v)
	// if err == nil {
	// 	t.Errorf(
	// 		"testRedBlackTreeInsert error, Repeated insertion without error, index: %d %+v", v, err1)
	// }
}

func testRedBlackTreeLength(t *testing.T, tmp *RedBlackTree, v int) {
	if tmp.Length() != v {
		t.Errorf("testRedBlackTreeLength error, want :%d, get: %d", v, tmp.Length())
	}
}

func testRedBlackTreeDepth(t *testing.T, tmp *RedBlackTree, v int) {
	d := tmp.Depth()
	if d != v {
		t.Errorf("testRedBlackTreeDepth error, want :%d, get: %d, root: %+v", v, d, tmp.root)
	}
}

func testPrevOrderVist(t *testing.T, tmp *RedBlackTree, want []int) {
	r := new(link.ArrayLinkList).Init(10)
	tmp.PrevOrderVist(resultSave(r))
	for i := 0; i < r.Length(); i++ {
		if r.Get(i).(int) != want[i] {
			t.Errorf("testPrevOrderVist, want %d: %d, but get %d", i, want[i], r.Get(i).(int))
		}
	}
}

func testInOrderVist(t *testing.T, tmp *RedBlackTree, want []int) {
	r := new(link.ArrayLinkList).Init(10)
	tmp.InOrderVist(resultSave(r))
	for i := 0; i < r.Length(); i++ {
		if r.Get(i).(int) != want[i] {
			t.Errorf(
				"testPrevOrderVist, want %d: %d, but get %d",
				i, want[i], r.Get(i).(int))
		}
	}
}

func testPostOrderVist(t *testing.T, tmp *RedBlackTree, want []int) {
	r := new(link.ArrayLinkList).Init(10)
	tmp.PostOrderVist(resultSave(r))
	for i := 0; i < r.Length(); i++ {
		if r.Get(i).(int) != want[i] || r.Length() != len((want)) {
			t.Errorf(
				"%s, want %d: %d, get %d, result length: %d, want length: %d",
				"testPrevOrderVist",
				i, want[i], r.Get(i).(int), r.Length(), len(want))
		}
	}
}

func testRedBlackTreeDelete(t *testing.T, tmp *RedBlackTree, i int, want []int) {
	// r := new(link.ArrayLinkList).Init(10)
	tmp.Delete(i)
	// tmp.InOrderVist(resultSave(r))
	// for i := 0; i < r.Length(); i++ {
	// if r.Get(i).(int) != want[i] || r.Length() != len((want)) {
	// 	t.Errorf(
	// 		"%s, want %d: %d, get %d, result length: %d, want length: %d, %+v",
	// 		"testPrevOrderVist",
	// 		i, want[i], r.Get(i).(int), r.Length(), len(want), tmp.root)
	// }
	// }
}
