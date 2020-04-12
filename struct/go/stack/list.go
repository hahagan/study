// 两种stack实现，数组线性表与链表
package stack

import (
	"github.com/hahagan/study/struct/go/list/pointer"
)

// 底层使用指针链表作为存储
type ListStack struct {
	pointer.PointerList
}

func (s *ListStack) Init() *ListStack {
	s.PointerList.Init()
	return s
}

func (s *ListStack) GetTop() interface{} {
	return s.Get(0)
}

func (s *ListStack) Pop() interface{} {
	r := s.GetTop()
	s.Delete(0)
	return r
}

func (s *ListStack) Push(i interface{}) {
	s.Insert(0, i)
}
