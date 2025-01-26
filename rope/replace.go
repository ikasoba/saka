package rope

func Replace(n RopeNode, offset int, length int, replacement string) RopeNode {
	l, tmp := n.Split(offset)
	_, r := tmp.Split(length)

	return Concat(l.Insert(offset, replacement), r).Rebalance()
}
