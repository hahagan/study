package queue

import (
	// "fmt"
	// "reflect"
	"testing"
)

func TestArrayQueue(t *testing.T) {
	s := new(ArrayQueue)
	s.Init(4)
	for i := 0; i < 20; i++ {
		testArrayQueuePush(t, s, i)
	}
	testArrayQueueGetTop(t, s, 0)
	for i := 0; i < 20; i++ {
		testArrayQueuePop(t, s, i)
	}

	if item, err := s.Pop(); item != nil || err == nil {
		t.Errorf("Test ArrayQueue Pop() error, item: %d", err)
	}
	testArrayQueueDestroy(t, s)
	s.Init(4)
	for i := 0; i < 4; i++ {
		testArrayQueuePush(t, s, i)

	}
	for i := 0; i < 3; i++ {
		testArrayQueuePop(t, s, i)

	}
	for i := 0; i < 10; i++ {
		testArrayQueuePush(t, s, i)

	}
	testArrayQueueClear(t, s)
}

func testArrayQueueDestroy(t *testing.T, l *ArrayQueue) {
	l.Destroy()
	if l.length != 0 || l.head != -1 || l.end != -1 {
		t.Error("Test ArrayQueue Destroy() error")
	}
}

func testArrayQueueClear(t *testing.T, l *ArrayQueue) {
	l.Clear()
	if l.length != 0 || l.head != -1 || l.end != -1 {
		t.Error("Test ArrayQueue Clear() error")
	}
}

func testArrayQueueGetTop(t *testing.T, s *ArrayQueue, want int) {
	i := s.GetHead()
	if i != want {
		t.Errorf("Test ArrayQueue GetTop() error, want: %d, get: %d", want, i)
	}
}

func testArrayQueuePop(t *testing.T, s *ArrayQueue, want interface{}) {
	lengthOld := s.Length()
	i, _ := s.Pop()
	lengthNew := s.Length()
	if i != want {
		t.Errorf("Test ArrayQueue Pop() error, want: %d, get: %d", want, i)
	}
	if lengthNew != lengthOld-1 {
		t.Errorf("Test ArrayQueue Pop() error, after pop %d, length: %d, before pop length: %d", want, lengthNew, lengthOld)
	}
}

func testArrayQueuePush(t *testing.T, s *ArrayQueue, i interface{}) {
	lengthOld := s.Length()
	s.Push(i)
	// want := s.GetHead()
	lengthNew := s.Length()
	// if i != want {
	// 	t.Errorf("Test ArrayQueue Push() error, want: %d, get: %d", i, want)
	// }
	if lengthNew != lengthOld+1 {
		t.Errorf("Test ArrayQueue Push() error, after push %d, length: %d, before push length: %d", i, lengthNew, lengthOld)
	}
}
