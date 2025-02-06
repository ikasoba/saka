package rope

func (*RopeEmptyNode) Runes() []rune {
	return []rune{}
}

func (n *RopeLeafNode) Runes() []rune {
	return n.text
}

func (n *RopeParentNode) Runes() []rune {
	return append(append([]rune{}, n.left.Runes()...), n.right.Runes()...)
}
