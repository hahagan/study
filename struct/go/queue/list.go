// 两种stack实现，数组线性表与链表
package queue

import (
	"fmt"

	"github.com/hahagan/study/struct/go/list/pointer"
)

// 底层使用数组(静态链表)作为存储
type ListQueue struct {
	pointer.DoubleList
	end    *pointer.DoubleList
	head   *pointer.DoubleList
	length int
}

func (s *ListQueue) Init() *ListQueue {
	s.head = s.DoubleList.Init()
	s.end = s.head
	return s
}

func (s *ListQueue) Destroy() {
	s.DoubleList.Destroy()
	s.head = nil
	s.end = nil
}
func (s *ListQueue) Clear() {
	s.DoubleList.Clear()
	s.head = nil
	s.end = nil
}

func (s *ListQueue) GetHead() interface{} {

	return s.DoubleList.Get(0)
}

// 返回对头并删除对头元素
func (s *ListQueue) Pop() (interface{}, error) {
	if s.DoubleList.Length() == 0 {
		return nil, fmt.Errorf("out of queue")
	}

	r := s.GetHead()
	s.DoubleList.Delete(0)
	return r, nil
}

func (s *ListQueue) Length() int {
	return s.DoubleList.Length()
}

// 插入队尾
func (s *ListQueue) Push(i interface{}) {
	s.end.Insert(1, i)
	s.end = s.end.Next
}
