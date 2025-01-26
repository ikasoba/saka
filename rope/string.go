package rope

func (*RopeEmptyNode) String() string {
	return ""
}

func (n *RopeLeafNode) String() string {
	return string(n.text)
}

func (n *RopeParentNode) String() string {
	return n.left.String() + n.right.String()
}
