package rope

func (n *RopeEmptyNode) GetLineNumber(offset int) int {
	return 0
}

func (n *RopeLeafNode) GetLineNumber(offset int) int {
	return countLF(n.text[:offset])
}

func (n *RopeParentNode) GetLineNumber(offset int) int {
	if offset < n.length {
		return n.left.GetLineNumber(offset)
	} else {
		return getLineBreaks(n.left) + n.right.GetLineNumber(offset-n.length)
	}
}

func (n *RopeEmptyNode) GetLineStart(line int) int {
	return 0
}

func (n *RopeLeafNode) GetLineStart(line int) int {
	ln := 0
	i := 0
	lastLineStart := 0

	for _, x := range n.text {
		if ln == line {
			return i
		}

		if x == '\n' {
			ln += 1

			lastLineStart = i + 1
		}

		i++
	}

	return lastLineStart
}

func (n *RopeParentNode) GetLineStart(line int) int {
	if line <= getLineBreaks(n.left) {
		return n.left.GetLineStart(line)
	} else {
		res := n.right.GetLineStart(line - getLineBreaks(n.left))
		if res > 0 {
			return n.length + res
		} else {
			return n.left.GetLineStart(getLineBreaks(n.left))
		}
	}
}

func (n *RopeEmptyNode) GetLineEnd(line int) int {
	return 0
}

func (n *RopeLeafNode) GetLineEnd(line int) int {
	ln := 0
	i := 0

	for _, x := range n.text {
		if x == '\n' {
			if ln == line {
				return i + 1
			}

			ln += 1
		}

		i++
	}

	return i
}

func (n *RopeParentNode) GetLineEnd(line int) int {
	if line <= getLineBreaks(n.left) {
		return n.left.GetLineEnd(line)
	} else {
		res := n.right.GetLineEnd(line - getLineBreaks(n.left))
		if res > 0 {
			return n.length + res
		} else {
			return n.left.GetLineEnd(getLineBreaks(n.left))
		}
	}
}

func GetIndexFromRowCol(n RopeNode, row int, col int) int {
	return n.GetLineStart(row) + col
}
