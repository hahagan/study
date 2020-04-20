package tree

import (
	"fmt"

	_ "github.com/hahagan/study/struct/go/list/link"
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

func (t *RedBlackTree) leftRotate(z *RedBlackTreeNode) {
	y := z.right
	z.right = y.left

	if y.left != t.leaf {
		y.left.parent = z
	}

	y.parent = z.parent

	if z == t.root {
		t.root = y
	} else if z == z.parent.left {
		z.parent.left = y
	} else {
		z.parent.right = y
	}
	y.left = z
	z.parent = y

}

func (t *RedBlackTree) rightRotate(z *RedBlackTreeNode) {
	y := z.left
	z.left = y.right

	if y.right != t.leaf {
		y.right.parent = z
	}

	y.parent = z.parent
	if z == t.root {
		t.root = y
	} else if z.parent.left == z {
		z.parent.left = y
	} else {
		z.parent.right = y
	}
	y.right = z
	z.parent = y
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
		value:  v,
		index:  i,
		color:  false,
		left:   t.leaf,
		right:  t.leaf,
		parent: t.leaf,
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

	if p == t.leaf {
		node.parent = t.leaf
		t.root = node
	} else if p.index < node.index {
		p.right = node
	} else {
		p.left = node
	}
	node.parent = p

	t.insertFix(node)
	t.length++
	return nil
}

func (t *RedBlackTree) insertFix(s *RedBlackTreeNode) {
	if s.parent == t.leaf {
		s.color = true
		return
	}
	for !s.parent.color {
		if s.parent == s.parent.parent.left {
			uncle := s.parent.parent.right
			if !uncle.color {
				uncle.color = true
				s.parent.color = true
				s.parent.parent.color = false
				s = s.parent.parent
			} else {
				if s == s.parent.right {
					s = s.parent
					t.leftRotate(s)
				}
				s.parent.color = true
				s.parent.parent.color = false
				t.rightRotate(s.parent.parent)
			}
		} else {
			if s.parent == s.parent.parent.right {
				uncle := s.parent.parent.left
				if !uncle.color {
					uncle.color = true
					s.parent.color = true
					s.parent.parent.color = false
					s = s.parent.parent
				} else {
					if s == s.parent.left {
						s = s.parent
						t.rightRotate(s)
					}
					s.parent.color = true
					s.parent.parent.color = false
					t.leftRotate(s.parent.parent)
				}
			}
		}

	}

	t.root.color = true
}

func (t *RedBlackTree) locateNode(i int) (*RedBlackTreeNode, error) {
	cur := t.root
	if cur == t.leaf {
		return nil, fmt.Errorf("Can't localte leaf")
	}

	if cur.index == i {
		return cur, nil
	} else if cur.index < i {
		cur = cur.right
	} else {
		cur = cur.left
	}

	return nil, fmt.Errorf("Can't localte leaf")
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
	for cur.left != t.leaf {
		cur = cur.left
	}
	return cur
}

func (t *RedBlackTree) delete(z *RedBlackTreeNode) error {
	if z == t.leaf {
		return fmt.Errorf("Can't delete leaf")
	}
	y := z
	color := y.color
	x := t.leaf
	if z.left == t.leaf {
		fmt.Println("-------------------")

		x = z.right
		t.instead(z, z.right)
	} else if z.right == t.leaf {
		fmt.Println("qqqqqqqqqqqqqqqqqqqqqq")

		x = z.left
		t.instead(z, z.left)
	} else {
		fmt.Println("wwwwwwwwwwwwwwwwwwwwwwww", t.length)

		y = t.minNode(z.right)
		color = y.color

		fmt.Println("wwwwwwwwwwwwwwwwwwwwwwww")
		x = y.right
		if y.parent == z {
			x.parent = y
		} else {
			t.instead(y, y.right)
			y.right = z.right
			y.right.parent = y
		}
		t.instead(z, y)
		y.left = z.left
		z.left.parent = y
		y.color = z.color

	}

	if color {
		t.deleteFix(x)
	}
	return nil
}

func (t *RedBlackTree) deleteFix(x *RedBlackTreeNode) {
	for x != t.root && x.color {
		if x == x.parent.left {
			w := x.parent.right
			if !w.color {
				x.parent.color = false
				w.color = true
				t.leftRotate(x.parent)
				w = x.parent.parent
				fmt.Println("111111111111111111")
			}

			if w.left.color && w.right.color {
				w.color = false
				x = x.parent
				fmt.Println("222222222222222222")
			} else {
				if w.right.color {
					w.left.color = true
					w.color = false
					t.rightRotate(w)
					w = x.parent.right
					fmt.Println("33333333333333333333")
				}
				w.color = w.parent.color
				w.parent.color = true
				w.right.color = true
				t.leftRotate(x.parent)
				x = t.root
				fmt.Println("4444444444444444444444")
			}
		} else {
			w := x.parent.left
			if !w.color {
				x.parent.color = false
				w.color = true
				t.rightRotate(x.parent)
				w = x.parent.left
				fmt.Println("-11111111111111111")
			}
			if w.left.color && w.right.color {
				w.color = false
				x = x.parent
				fmt.Println("-22222222222222222")
			} else {
				if w.left.color {
					w.right.color = true
					w.color = false
					t.leftRotate(w)
					w = x.parent.left
					fmt.Println("-33333333333333333333")
				}
				w.color = x.parent.color
				w.parent.color = true
				w.left.color = true
				t.rightRotate(x.parent)
				x = t.root
				fmt.Println("-44444444444444444444444444")
			}
		}
	}
	x.color = true
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
