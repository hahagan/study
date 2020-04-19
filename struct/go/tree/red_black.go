package tree

import (
	"fmt"

	"github.com/hahagan/study/struct/go/queue"
	"github.com/hahagan/study/struct/go/stack"
)

type RedBlackTreeNode struct {
	parent *RedBlackTreeNode
	left   *RedBlackTreeNode
	right  *RedBlackTreeNode
	index  int
	value  interface{}
	color  bool
}

type RedBlackTree struct {
	root   *RedBlackTreeNode
	leaf   *RedBlackTreeNode
	length int
}

func (t *RedBlackTree) leftRotate(x *RedBlackTreeNode) {
	y := x.right
	x.right = y.left

	if y.left != nil {
		y.left.parent = x
	}

	y.parent = x.parent
	if x.parent == nil {
		t.root = y
	} else if x == x.parent.left {
		x.parent.left = y
	} else {
		x.parent.right = y
	}
	y.left = x
	x.parent = y

}

func (t *RedBlackTree) rightRotate(y *RedBlackTreeNode) {
	x := y.left
	y.left = x.right

	if x.right != nil {
		x.right.parent = y
	}

	x.parent = y.parent
	if y.parent == nil {
		t.root = x
	} else if y.parent.left == y {
		y.parent.left = x
	} else {
		y.parent.right = y
	}
	x.right = y
	y.parent = x
}

func (t *RedBlackTree) Init() *RedBlackTree {
	t.length = 0
	t.leaf = &RedBlackTreeNode{
		color: true,
	}
	t.root = t.leaf
	return t
}

func (t *RedBlackTree) Length() int {
	return t.length
}

func (t *RedBlackTree) Insert(i int, v interface{}) error {
	cur := t.root
	p := cur
	node := &RedBlackTreeNode{
		value: v,
		index: i,
		color: false,
		left:  t.leaf,
		right: t.leaf,
	}

	for cur != t.leaf {
		p = cur
		if p.index < node.index {
			cur = p.right
		} else if p.index > node.index {
			cur = p.left
		} else {
			return fmt.Errorf("index %d has exiting", i)
		}
	}

	node.parent = p

	if p == t.leaf {
		node.parent = t.leaf
		t.root = node
	} else if p.index < node.index {
		p.right = node
	} else {
		p.left = node
	}

	t.insertFix(node)
	t.length++
	return nil
}

func (t *RedBlackTree) insertFix(x *RedBlackTreeNode) {
	if x.parent == t.leaf {
		x.color = true
		return
	}
	for !x.parent.color {
		if x.parent == x.parent.parent.left {
			uncle := x.parent.parent.right
			if !uncle.color {
				uncle.color = true
				x.parent.color = true
				x.parent.parent.color = false
				x = x.parent.parent
			} else if x == x.parent.right {
				x = x.parent
				t.leftRotate(x)
			} else {
				x.parent.color = true
				x.parent.parent.color = false
				t.rightRotate(x.parent.parent)
			}
		} else {
			uncle := x.parent.parent.right
			if !uncle.color {
				uncle.color = true
				x.parent.color = true
				x.parent.parent.color = false
				x = x.parent.parent
			} else if x == x.parent.left {
				x = x.parent
				t.rightRotate(x)
			} else {
				x.parent.color = true
				x.parent.parent.color = false
				t.leftRotate(x.parent.parent)
			}
		}
	}

}

func (t *RedBlackTree) locateNode(i int) (*RedBlackTreeNode, error) {
	cur := t.root
	p := cur
	for cur != t.leaf {
		p = cur
		if p.index < i {
			cur = p.right
		} else if p.index > i {
			cur = p.left
		} else {
			break
		}
	}

	if p == t.leaf {
		return nil, fmt.Errorf("Can't delete leaf")
	}
	return p, nil
}

func (t *RedBlackTree) Delete(i int) error {
	p, err := t.locateNode(i)
	if err != nil {
		return err
	}

	err = t.delete(p)
	if err == nil {
		t.length--
	}
	return nil
}

func (t *RedBlackTree) instead(x, y *RedBlackTreeNode) {
	if x.parent == t.leaf {
		t.root = y
	} else if x.parent.left == x {
		x.parent.left = y
	} else {
		x.parent.right = y
	}
	y.parent = x.parent
}

func (t *RedBlackTree) minNode(x *RedBlackTreeNode) *RedBlackTreeNode {
	cur := x
	p := cur
	for cur != t.leaf {
		p = cur
		cur = x.left
	}
	return p
}

func (t *RedBlackTree) delete(x *RedBlackTreeNode) error {
	if x == t.leaf {
		return fmt.Errorf("Can't delete leaf")
	}
	color := x.color
	var y *RedBlackTreeNode
	if x.left == t.leaf {
		y = x.right
		t.instead(x, y)
	} else if x.right == t.leaf {
		y = x.left
		t.instead(x, y)
	} else {
		d := t.minNode(x.right)
		y = d.right
		color = d.color
		if d.parent != x {
			t.instead(d, d.right)
			d.right = x.right
			x.right.parent = d
		}
		t.instead(x, d)
		d.left = x.left
		x.left.parent = d
		d.color = x.color

	}

	if color {
		t.deleteFix(y)
	}
	return nil
}

func (t *RedBlackTree) deleteFix(x *RedBlackTreeNode) {
	for x != t.leaf && x.color {
		if x == x.parent.left {
			w := x.parent.right
			if !w.color {
				x.parent.color = false
				w.color = false
				t.leftRotate(x.parent)
			} else if w.left.color && w.right.color && w.color {
				w.color = false
				x = x.parent
			} else if !w.left.color && w.right.color && w.color {
				w.left.color = true
				w.color = false
				t.rightRotate(w)
			} else if !w.right.color && w.color {
				w.color = x.parent.color
				x.parent.color = true
				w.right.color = true
				t.leftRotate(x.parent)
				x = t.leaf
			}
		}
	}
}

func (t *RedBlackTree) Depth() int {
	depth := 0
	if t.root == t.leaf {
		return depth
	}
	q := new(queue.ListQueue).Init()
	q.Push(t.root)
	q1 := new(queue.ListQueue).Init()
	for {
		for q.Length() > 0 {
			item, _ := q.Pop()
			cur := item.(*RedBlackTreeNode)
			if cur.left != t.leaf {
				q1.Push(cur.left)
			}
			if cur.right != t.leaf {
				q1.Push(cur.left)
			}
		}
		depth++
		if q1.Length() != 0 {
			tmp := q
			q = q1
			q1 = tmp
		} else {
			break
		}
	}

	return depth

}

func (t *RedBlackTree) PrevOrderVist(f func(interface{}) error) error {
	s := new(stack.ListStack)
	s.Init()
	if t.root != t.leaf {
		s.Push(t.root)

	}

	for s.Length() > 0 {
		cur, ok := s.Pop().(*RedBlackTreeNode)
		if !ok {
			return fmt.Errorf("Convert stack item to *RedBlackTree failed")
		}
		if err1 := f(cur.value); err1 != nil {
			return err1
		}
		if cur.right != t.leaf {
			s.Push(cur.right)
		}
		if cur.left != t.leaf {
			s.Push(cur.left)
		}

	}

	return nil
}

func (t *RedBlackTree) InOrderVist(f func(interface{}) error) error {
	s := new(stack.ListStack)
	s.Init()
	cur := t.root
	for cur != t.leaf {
		s.Push(cur)
		cur = cur.left
	}

	for s.Length() > 0 {
		var ok bool
		cur, ok = s.Pop().(*RedBlackTreeNode)
		if !ok {
			return fmt.Errorf("Convert stack item to *RedBlackTree failed")
		}
		if err1 := f(cur.value); err1 != nil {
			return err1
		}

		if cur.right != t.leaf {
			cur = cur.right
			for cur != t.leaf {
				s.Push(cur)
				cur = cur.left
			}
		}
	}

	return nil
}

func (t *RedBlackTree) PostOrderVist(f func(interface{}) error) error {
	s := new(stack.ListStack)
	s.Init()
	cur := t.root
	for cur != t.leaf {
		s.Push(cur)
		if cur.left != t.leaf {
			cur = cur.left
		} else {
			cur = cur.right
		}
	}

	for s.Length() > 0 {
		var ok bool
		cur, ok = s.Pop().(*RedBlackTreeNode)
		if !ok {
			return fmt.Errorf("Convert stack item to *RedBlackTree failed")
		}
		if err1 := f(cur.value); err1 != nil {
			return err1
		}

		if s.Length() > 0 {
			var p *RedBlackTreeNode
			p, ok = s.GetTop().(*RedBlackTreeNode)
			if !ok {
				return fmt.Errorf("Convert stack item to *RedBlackTree failed")
			}
			if p.left == cur {
				cur = p.right
				for cur != t.leaf {
					s.Push(cur)
					if cur.left != t.leaf {
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
