package queue

import (
	"testing"
)

func TestListQueue(t *testing.T) {
	s := new(ListQueue)
	s.Init()
	for i := 0; i < 20; i++ {
		testListQueuePush(t, s, i)
	}
	testListQueueGetTop(t, s, 0)

	for i := 0; i < 20; i++ {
		testListQueuePop(t, s, i)

	}

	if item, err := s.Pop(); item != nil || err == nil {
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
	// lengthOld := s.Length()
	s.Push(i)
	// want := s.GetHead()
	// lengthNew := s.Length()
	// // if i != want {
	// // 	t.Errorf("Test ListQueue Push() error, want: %d, get: %d", i, want)
	// // }
	// if lengthNew != lengthOld+1 {
	// 	t.Errorf("Test ListQueue Push() error, after push %d, length: %d, before push length: %d", i, lengthNew, lengthOld)
	// }
}
