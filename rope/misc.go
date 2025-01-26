package rope

func countLF(runes []rune) int {
	n := 0

	for _, x := range runes {
		if x == '\n' {
			n += 1
		}
	}

	return n
}

func getLineBreaks(n RopeNode) int {
	switch n := n.(type) {
	case *RopeParentNode:
		return n.LineBreaks

	case *RopeLeafNode:
		return n.LineBreaks

	default:
		return 0
	}
}

func GetLineBreaks(n RopeNode) int {
	switch n := n.(type) {
	case *RopeParentNode:
		return n.LineBreaks

	case *RopeLeafNode:
		return n.LineBreaks

	default:
		return 0
	}
}
