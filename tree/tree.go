package tree

type Tree[T any] struct {
	Value    T
	HasValue bool
	Left     *Tree[T]
	Right    *Tree[T]
}

func (n *Tree[T]) Find(cond func(val T) int) *Tree[T] {
	switch cond(n.Value) {
	case -1:
		if n.Left == nil {
			return n
		} else {
			return n.Left.Find(cond)
		}

	case 1:
		if n.Right == nil {
			return n
		} else {
			return n.Right.Find(cond)
		}

	default:
		return n
	}
}

func (n *Tree[T]) Put(cond func(a T, b T) int, node *Tree[T]) *Tree[T] {
	if !n.HasValue {
		n.HasValue = node.HasValue
		n.Value = node.Value
		n.Left = node.Left
		n.Right = node.Right

		return n
	}

	switch cond(node.Value, n.Value) {
	case -1:
		if n.Left == nil {
			n.Left = node

			return n
		}

		n.Left.Put(cond, node)

		return n

	case 1:
		if n.Right == nil {
			n.Right = node

			return n
		}

		n.Right.Put(cond, node)

		return n

	default:
		n.HasValue = node.HasValue
		n.Value = node.Value

		if node.Left != nil && n.Left != nil {
			n.Left = n.Left.Put(cond, node.Left)
		} else if node.Left != nil && n.Left == nil {
			n.Left = node.Left
		}

		if node.Right != nil && n.Right != nil {
			n.Right = n.Right.Put(cond, node.Right)
		} else if node.Right != nil && n.Right == nil {
			n.Right = node.Right
		}

		return n
	}
}

func (n *Tree[T]) Remove(cond func(val T) int, cond2 func(a T, b T) int) *Tree[T] {
	node := n.Find(cond)

	if node.Right != nil {
		node.HasValue = node.Right.HasValue
		node.Value = node.Right.Value
		node.Left = node.Right.Left
		node.Right = node.Right.Right

		if node.Left != nil {
			node.Put(cond2, node.Left)
		}
	} else if node.Left != nil {
		node.HasValue = node.Left.HasValue
		node.Value = node.Left.Value
		node.Left = node.Left.Left
		node.Right = node.Left.Right
	}

	return n
}
