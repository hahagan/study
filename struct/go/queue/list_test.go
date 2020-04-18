package queue

import (
	"fmt"
	"testing"
)

func TestListQueue(t *testing.T) {
	fmt.Println("")
	s := new(ListQueue)
	s.Init()
	times := 100
	for i := 0; i < times; i++ {
		testListQueuePush(t, s, i)
	}
	testListQueueGetTop(t, s, 0)

	for i := 0; i < times; i++ {
		testListQueuePop(t, s, i)

	}

	if item, err := s.Pop(); item != nil || err == nil {
		t.Errorf("Test ListQueue Pop() error, item: %d", item)
	}

	testListQueuePush(t, s, 0)
	testListQueuePop(t, s, 0)
	testListQueuePush(t, s, 1)
	testListQueuePop(t, s, 1)

	testListQueueClear(t, s)
	testListQueueDestroy(t, s)
}

func TestLisQueueEndFunction(t *testing.T) {
	s := new(ListQueue)
	s.Init()
	for i := 0; i < 20; i++ {
		testListQueuePushHead(t, s, i)
	}
	testListQueueGetEnd(t, s, 0)

	for i := 0; i < 20; i++ {
		testListQueuePopEnd(t, s, i)

	}

	if item, err := s.PopEnd(); item != nil || err == nil {
		t.Errorf("Test ListQueue Pop() error, item: %d", item)
	}

	testListQueuePushHead(t, s, 0)
	testListQueuePopEnd(t, s, 0)
	testListQueuePushHead(t, s, 1)
	testListQueuePopEnd(t, s, 1)

	testListQueueClear(t, s)
	testListQueueDestroy(t, s)
}

func TestLisQueueMix(t *testing.T) {
	s := new(ListQueue)
	s.Init()

	// 2
	testListQueuePush(t, s, 2)
	// 1,2
	testListQueuePushHead(t, s, 1)
	// 0,1,2
	testListQueuePushHead(t, s, 0)
	// 0,1,2,3
	testListQueuePush(t, s, 3)

	// 1,2,3
	testListQueuePop(t, s, 0)
	// 1,2
	testListQueuePopEnd(t, s, 3)
	// 1
	testListQueuePopEnd(t, s, 2)
	// nil
	testListQueuePopEnd(t, s, 1)

	if item, err := s.PopEnd(); item != nil || err == nil {
		t.Errorf("Test ListQueue Pop() error, item: %d", item)
	}

	testListQueueClear(t, s)
	testListQueueDestroy(t, s)
}

func testListQueueDestroy(t *testing.T, l *ListQueue) {
	l.Destroy()
	if l.Length() != -1 || l.head != nil || l.end != nil {
		t.Error("Test ListQueue Destroy() error")
	}
}

func testListQueueClear(t *testing.T, l *ListQueue) {
	l.Clear()
	if l.Length() != 0 || l.head != nil || l.end != nil {
		t.Error("Test ListQueue Clear() error")
	}
}

func testListQueueGetTop(t *testing.T, s *ListQueue, want int) {
	i := s.GetHead()
	if i != want {
		t.Errorf("Test ListQueue GetTop() error, want: %d, get: %d", want, i)
	}
}

func testListQueuePop(t *testing.T, s *ListQueue, want interface{}) {

	i, _ := s.Pop()
	if i != want {
		t.Errorf("Test ListQueue Pop() error, want: %d, get: %d", want, i)
		t.Error(s.head, s.head.Length())
		for cur := s.head; cur.Next != nil; cur = cur.Next {
			t.Error("|------item: ", cur.Next, cur.Prev)

		}
	}

}

func testListQueuePush(t *testing.T, s *ListQueue, i interface{}) {
	s.Push(i)
}

func testListQueueGetEnd(t *testing.T, s *ListQueue, want int) {
	i := s.GetEnd()
	if i != want {
		t.Errorf("Test ListQueue GetEnd() error, want: %d, get: %d", want, i)

	}
}

func testListQueuePopEnd(t *testing.T, s *ListQueue, want interface{}) {

	i, _ := s.PopEnd()
	if i != want {
		t.Errorf("Test ListQueue PopEnd() error, want: %d, get: %d", want, i)
		t.Error(s.head, s.head.Length())
		for cur := s.head; cur.Next != nil; cur = cur.Next {
			t.Error("|------item: ", cur.Next, cur.Prev, s.end)

		}
	}

}

func testListQueuePushHead(t *testing.T, s *ListQueue, i interface{}) {
	s.PushHead(i)
}
