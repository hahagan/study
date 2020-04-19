package tree

import (
	"fmt"

	"github.com/hahagan/study/struct/go/stack"
)

type BinaryTreeNode struct {
	parent *BinaryTreeNode
	left   *BinaryTreeNode
	right  *BinaryTreeNode
	value  interface{}
	index  int
}

func (t *BinaryTreeNode) Insert(i int, node *BinaryTreeNode) {
	cur := t
	for cur != nil {
		if cur.index <= node.index {
			tmp := cur.right
			if tmp == nil {
				cur.right = node
				node.parent = cur
				break
			} else {
				cur = tmp
			}
		} else if cur.index > node.index {
			tmp := cur.left
			if tmp == nil {
				cur.left = node
				node.parent = cur
				break
			} else {
				cur = tmp
			}
		}
	}
}

func (n *BinaryTreeNode) Depth() int {
	max := 1
	left := 0
	if n.left != nil {
		left = n.left.Depth() + 1
	}
	right := 0
	if n.right != nil {
		right = n.right.Depth() + 1
	}
	if left > max {
		max = left
	}
	if right > max {
		max = right
	}
	return max
}

func (n *BinaryTreeNode) PrevOrderVist(f func(interface{}) error) error {
	s := new(stack.ListStack)
	s.Init()
	s.Push(n)

	for s.Length() > 0 {
		cur, ok := s.Pop().(*BinaryTreeNode)
		if !ok {
			return fmt.Errorf("Convert stack item to *BinaryTreeNode failed")
		}
		if err1 := f(cur.value); err1 != nil {
			return err1
		}
		if cur.right != nil {
			s.Push(cur.right)
		}
		if cur.left != nil {
			s.Push(cur.left)
		}

	}

	return nil
}

func (n *BinaryTreeNode) InOrderVist(f func(interface{}) error) error {
	s := new(stack.ListStack)
	s.Init()
	cur := n
	for cur != nil {
		s.Push(cur)
		cur = cur.left
	}

	for s.Length() > 0 {
		var ok bool
		cur, ok = s.Pop().(*BinaryTreeNode)
		if !ok {
			return fmt.Errorf("Convert stack item to *BinaryTreeNode failed")
		}
		if err1 := f(cur.value); err1 != nil {
			return err1
		}

		if cur.right != nil {
			cur = cur.right
			for cur != nil {
				s.Push(cur)
				cur = cur.left
			}
		}
	}

	return nil
}

func (n *BinaryTreeNode) PostOrderVist(f func(interface{}) error) error {
	s := new(stack.ListStack)
	s.Init()
	cur := n
	for cur != nil {
		s.Push(cur)
		if cur.left != nil {
			cur = cur.left
		} else {
			cur = cur.right
		}
	}

	for s.Length() > 0 {
		var ok bool
		cur, ok = s.Pop().(*BinaryTreeNode)
		if !ok {
			return fmt.Errorf("Convert stack item to *BinaryTreeNode failed")
		}
		if err1 := f(cur.value); err1 != nil {
			return err1
		}

		if s.Length() > 0 {
			var p *BinaryTreeNode
			p, ok = s.GetTop().(*BinaryTreeNode)
			if !ok {
				return fmt.Errorf("Convert stack item to *BinaryTreeNode failed")
			}
			if p.left == cur {
				cur = p.right
				for cur != nil {
					s.Push(cur)
					if cur.left != nil {
						cur = cur.left
					} else {
						cur = cur.right
					}
				}
			}

		}
	}

	return nil
}

type BinaryTree struct {
	root   *BinaryTreeNode
	length int
}

func (t *BinaryTree) Init() *BinaryTree {
	t.length = 0
	t.root = nil
	return t
}

func (t *BinaryTree) Length() int {
	return t.length
}

func (t *BinaryTree) Depth() int {
	if t.root == nil {
		return 0
	}
	return t.root.Depth()
}

func (t *BinaryTree) Insert(i int, v interface{}) {
	cur := t.root
	node := &BinaryTreeNode{
		value: v,
		index: i,
	}

	if t.root == nil {
		t.root = node
	} else {
		cur.Insert(i, node)
	}

	t.length++

}

func (t *BinaryTree) PrevOrderVist(f func(interface{}) error) error {
	if t.root != nil {
		return t.root.PrevOrderVist(f)
	}
	return nil
}

func (t *BinaryTree) InOrderVist(f func(interface{}) error) error {
	if t.root != nil {
		return t.root.InOrderVist(f)
	}
	return nil
}

func (t *BinaryTree) PostOrderVist(f func(interface{}) error) error {
	if t.root != nil {
		return t.root.PostOrderVist(f)
	}
	return nil
}
