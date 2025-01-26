package rope

func (*RopeEmptyNode) Length() int {
	return 0
}

func (n *RopeLeafNode) Length() int {
	return len(n.text)
}

func (n *RopeParentNode) Length() int {
	return n.length + n.right.Length()
}
