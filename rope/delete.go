package rope

func (n *RopeEmptyNode) Delete(offset int, count int) RopeNode {
	return n
}

func (n *RopeLeafNode) Delete(offset int, count int) RopeNode {
	runes := append(n.text[:offset], n.text[offset+count:]...)

	return &RopeLeafNode{countLF(runes), runes}
}

func (n *RopeParentNode) Delete(offset int, count int) RopeNode {
	l, _ := n.Split(offset)
	_, r := n.Split(offset + count)

	return Concat(l, r).Rebalance()
}
