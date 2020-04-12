package stack

import (
	// "reflect"
	"testing"
)

func TestArrayStack(t *testing.T) {
	s := new(ArrayStack)
	s.Init(4)
	testArrayStackPush(t, s, 0)
	testArrayStackGetTop(t, s, 0)
	testArrayStackPop(t, s, 0)

}

func testArrayStackGetTop(t *testing.T, s *ArrayStack, want int) {
	i := s.GetTop()
	if i != want {
		t.Errorf("Test ArrayStack GetTop() error, want: %d, get: %d", want, i)
	}
}

func testArrayStackPop(t *testing.T, s *ArrayStack, want interface{}) {
	lengthOld := s.Length()
	i := s.Pop()
	lengthNew := s.Length()
	if i != want {
		t.Errorf("Test ArrayStack Pop() error, want: %d, get: %d", want, i)
	}
	if lengthNew != lengthOld-1 {
		t.Errorf("Test ArrayStack Pop() error, after pop length: %d, before pop length: %d", lengthNew, lengthOld)
	}
}

func testArrayStackPush(t *testing.T, s *ArrayStack, i interface{}) {
	lengthOld := s.Length()
	s.Push(i)
	want := s.GetTop()
	lengthNew := s.Length()
	if i != want {
		t.Errorf("Test ArrayStack Push() error, want: %d, get: %d", i, want)
	}
	if lengthNew != lengthOld+1 {
		t.Errorf("Test ArrayStack Push() error, after pop length: %d, before pop length: %d", lengthNew, lengthOld)
	}
}
