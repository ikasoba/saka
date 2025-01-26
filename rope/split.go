package rope

func (n *RopeEmptyNode) Split(offset int) (RopeNode, RopeNode) {
	return &RopeEmptyNode{}, &RopeEmptyNode{}
}

func (n *RopeLeafNode) Split(offset int) (RopeNode, RopeNode) {
	if len(n.text) == 0 {
		return &RopeEmptyNode{}, &RopeEmptyNode{}
	}

	l, r := n.text[:offset], n.text[offset:]

	return &RopeLeafNode{countLF(l), l}, &RopeLeafNode{countLF(r), r}
}

func (n *RopeParentNode) Split(offset int) (RopeNode, RopeNode) {
	if offset < n.length {
		l, r := n.left.Split(offset)

		return l.Rebalance(), RopeNode(&RopeParentNode{
			length:     r.Length(),
			LineBreaks: getLineBreaks(r) + getLineBreaks(n.right),
			left:       r,
			right:      n.right,
		}).Rebalance()
	} else if offset > n.length {
		l, r := n.right.Split(offset - n.length)

		return RopeNode(&RopeParentNode{
			length:     n.left.Length(),
			LineBreaks: getLineBreaks(n.left) + getLineBreaks(l),
			left:       n.left,
			right:      l,
		}).Rebalance(), r.Rebalance()
	} else {
		return n.left, n.right
	}
}
