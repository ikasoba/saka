package rope

import "io"

type RopeReader struct {
	stack []RopeNode
	index int
}

func NewReader(node RopeNode) *RopeReader {
	return &RopeReader{
		stack: []RopeNode{node},
		index: 0,
	}
}

func (r *RopeReader) ReadRune() (ch rune, size int, err error) {
loop:
	for {
		if len(r.stack) == 0 {
			return 0, 0, io.EOF
		}

		node := r.stack[len(r.stack)-1]

		switch node := node.(type) {
		case *RopeParentNode:
			r.stack = r.stack[:len(r.stack)-1]

			if node.right != nil {
				r.stack = append(r.stack, node.right)
			}

			if node.left != nil {
				r.stack = append(r.stack, node.left)
			}

			continue loop

		case *RopeEmptyNode:
			r.stack = r.stack[:len(r.stack)-1]
			continue loop

		case *RopeLeafNode:
			if r.index >= len(node.text) {
				r.stack = r.stack[:len(r.stack)-1]
				r.index = 0
				continue loop
			} else {
				ch := node.text[r.index]

				r.index += 1

				return ch, 1, nil
			}
		}

		break
	}

	return ch, size, err
}
