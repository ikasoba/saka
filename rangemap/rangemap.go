package rangemap

import (
	"github.com/ikasoba/saka/tree"
)

type Range[T any] struct {
	A   int
	B   int
	Val T
}

func CompareRange[T any](a Range[T], b Range[T]) int {
	if a.A >= b.A && a.B < b.B {
		return 0
	}

	if a.A <= b.A && a.B > b.B {
		return 0
	}

	if a.A >= b.A && a.B > b.B {
		return 1
	}

	if a.A <= b.A && a.B < b.B {
		return -1
	}

	if a.A < b.B {
		return -1
	}

	if a.A > b.B {
		return 1
	}

	return 0
}

type RangeMap[T any] struct {
	node tree.Tree[Range[T]]
}

func (m *RangeMap[T]) Get(a int, b int) (T, bool, Range[T]) {
	var tmp T

	res := m.node.Find(func(val Range[T]) int {
		return CompareRange(Range[T]{
			a, b, tmp,
		}, val)
	})

	if !res.HasValue {
		return tmp, false, res.Value
	}

	if CompareRange(Range[T]{a, b, tmp}, res.Value) != 0 {
		return tmp, false, res.Value
	}

	return res.Value.Val, true, res.Value
}

func (m *RangeMap[T]) Put(a int, b int, val T) {
	m.node.Put(CompareRange[T], &tree.Tree[Range[T]]{Value: Range[T]{
		a, b, val,
	}, HasValue: true})
}

func (m *RangeMap[T]) Remove(a int, b int) {
	var tmp T

	m.node.Remove(func(val Range[T]) int {
		return CompareRange(Range[T]{
			a, b, tmp,
		}, val)
	}, CompareRange[T])
}
