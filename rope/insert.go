package rope

import "errors"

func (n *RopeEmptyNode) Insert(offset int, text string) RopeNode {
	if offset != 0 {
		panic(errors.New("Out of range"))
	}

	runes := []rune(text)

	return &RopeLeafNode{countLF(runes), runes}
}

func (n *RopeLeafNode) Insert(offset int, text string) RopeNode {
	if offset > len(n.text) {
		panic(errors.New("Out of range"))
	}

	if offset == 0 {
		t := []rune(text)
		linebreaks := countLF(t)

		return &RopeLeafNode{n.LineBreaks + linebreaks, append(t, n.text...)}
	} else {
		lRunes := n.text[:offset]
		l := &RopeLeafNode{countLF(lRunes), lRunes}

		rRunes := append([]rune(text), n.text[offset:]...)
		r := &RopeLeafNode{countLF(rRunes), rRunes}

		return &RopeParentNode{
			length:     l.Length(),
			LineBreaks: l.LineBreaks + r.LineBreaks,
			left:       l,
			right:      r,
		}
	}
}

func (n *RopeParentNode) Insert(offset int, text string) RopeNode {
	l, r := n.Split(offset)

	runes := []rune(text)

	return Concat(Concat(l, &RopeLeafNode{countLF(runes), runes}), r).Rebalance()
}
