package rope

func Slice(n RopeNode, offset int, count int) RopeNode {
	if offset == count {
		return &RopeEmptyNode{}
	}

	_, r := n.Split(offset)
	res, _ := r.Split(count)

	return res
}
