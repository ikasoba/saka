package rope

const (
	splitThreshold = 128
	joinThreshold  = splitThreshold / 2
)

func (n *RopeEmptyNode) Rebalance() RopeNode {
	return n
}

func (n *RopeLeafNode) Rebalance() RopeNode {
	if len(n.text) > splitThreshold {
		offset := len(n.text) / 2

		lRunes, rRunes := n.text[:offset], n.text[offset:]
		l, r := &RopeLeafNode{countLF(lRunes), lRunes}, &RopeLeafNode{countLF(rRunes), rRunes}

		return &RopeParentNode{
			length:     l.Length(),
			LineBreaks: l.LineBreaks + r.LineBreaks,
			left:       l,
			right:      r,
		}
	} else {
		return n
	}
}

func (n *RopeParentNode) Rebalance() RopeNode {
	if n.length < joinThreshold || n.length > joinThreshold {
		return &RopeLeafNode{n.LineBreaks, n.Runes()}
	} else {
		l, r := n.left.Rebalance(), n.right.Rebalance()

		return &RopeParentNode{
			length:     l.Length(),
			LineBreaks: getLineBreaks(l) + getLineBreaks(r),
			left:       l,
			right:      r,
		}
	}
}
