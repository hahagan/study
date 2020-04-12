package stack

import (
	// "reflect"
	"testing"
)

func TestListStack(t *testing.T) {
	s := new(ListStack)
	s.Init()
	testListStackPush(t, s, 0)
	testListStackGetTop(t, s, 0)
	testListStackPop(t, s, 0)

}

func testListStackGetTop(t *testing.T, s *ListStack, want int) {
	i := s.GetTop()
	if i != want {
		t.Errorf("Test ListStack GetTop() error, want: %d, get: %d", want, i)
	}
}

func testListStackPop(t *testing.T, s *ListStack, want interface{}) {
	lengthOld := s.Length()
	i := s.Pop()
	lengthNew := s.Length()
	if i != want {
		t.Errorf("Test ListStack Pop() error, want: %d, get: %d", want, i)
	}
	if lengthNew != lengthOld-1 {
		t.Errorf("Test ListStack Pop() error, after pop length: %d, before pop length: %d", lengthNew, lengthOld)
	}
}

func testListStackPush(t *testing.T, s *ListStack, i interface{}) {
	lengthOld := s.Length()
	s.Push(i)
	want := s.GetTop()
	lengthNew := s.Length()
	if i != want {
		t.Errorf("Test ListStack Push() error, want: %d, get: %d", i, want)
	}
	if lengthNew != lengthOld+1 {
		t.Errorf("Test ListStack Push() error, after pop length: %d, before pop length: %d", lengthNew, lengthOld)
	}
}
