// 两种stack实现，数组线性表与链表
package queue

import (
	"fmt"

	"github.com/hahagan/study/struct/go/list/link"
)

// 底层使用数组(静态链表)作为存储
type ArrayQueue struct {
	link.ArrayLinkList
	end    int
	head   int
	length int
}

func (s *ArrayQueue) Init(capacity int) *ArrayQueue {
	s.ArrayLinkList.Init(capacity)
	s.head = -1
	s.end = -1
	s.length = 0
	return s
}

func (s *ArrayQueue) Destroy() {
	s.ArrayLinkList.Destroy()
	s.head = -1
	s.end = -1
	s.length = 0
}
func (s *ArrayQueue) Clear() {
	s.ArrayLinkList.Clear()
	s.head = -1
	s.end = -1
	s.length = 0
}

func (s *ArrayQueue) GetHead() interface{} {

	return s.Get(s.head)
}

// 返回对头并删除对头元素
func (s *ArrayQueue) Pop() (interface{}, error) {
	if s.length == 0 {
		return nil, fmt.Errorf("out of queue")
	}
	r := s.GetHead()
	s.head = (s.head + 1) % s.Capacity()
	s.length -= 1
	return r, nil
}

func (s *ArrayQueue) Length() int {
	return s.length
}

func (s *ArrayQueue) reverse(start int, end int) {
	j := end
	for i := start; i < j; i++ {
		tmp := s.Get(i)
		s.Set(i, s.Get(j))
		s.Set(j, tmp)
		j--
	}
}

func (s *ArrayQueue) leftRate(num int) {
	s.reverse(0, num-1)
	s.reverse(num, s.Capacity()-1)
	s.reverse(0, s.Capacity()-1)
}

// 插入队尾
func (s *ArrayQueue) Push(i interface{}) {
	length := s.Length()
	capacity := s.Capacity()
	var end int

	// 自动扩容
	if length == capacity {
		if s.end > s.head {
			s.leftRate(s.head)
		} else {
			s.leftRate(s.end)
		}
		s.head = 0
		s.end = length - 1
		end = length
	} else {
		end = (s.end + 1) % capacity
	}

	//插入
	if s.ArrayLinkList.Length() == s.length {
		s.Insert(end, i)
	} else {
		s.Set(end, i)
	}
	s.end = end
	if s.head == -1 {
		s.head = end
	}
	s.length += 1
}
