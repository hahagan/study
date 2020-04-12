// 两种stack实现，数组线性表与链表
package stack

import (
	"github.com/hahagan/study/struct/go/list/link"
)

// 底层使用数组作为存储
type ArrayStack struct {
	link.ArrayLinkList
}

func (s *ArrayStack) GetTop() interface{} {
	topIndex := s.ArrayLinkList.Length() - 1
	return s.ArrayLinkList.Get(topIndex)
}

func (s *ArrayStack) Pop() interface{} {
	topIndex := s.ArrayLinkList.Length() - 1
	r := s.ArrayLinkList.Get(topIndex)
	s.Delete(topIndex)
	return r
}

func (s *ArrayStack) Push(i interface{}) {
	topIndex := s.ArrayLinkList.Length()
	s.ArrayLinkList.Insert(topIndex, i)
}
