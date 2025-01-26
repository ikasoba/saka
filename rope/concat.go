package rope

func Concat(l RopeNode, r RopeNode) RopeNode {
	return &RopeParentNode{
		length:     l.Length(),
		LineBreaks: getLineBreaks(l) + getLineBreaks(r),
		left:       l,
		right:      r,
	}
}
