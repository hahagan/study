package tree

import (
	"fmt"
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

func (t *RedBlackTree) Insert(i int, v interface{}) error {
	cur := t.root
	var p *RedBlackTreeNode
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
	if x.parent == nil {
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
