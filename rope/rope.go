package rope

import "io"

type RopeNodeKind int

const (
	RopeKindEmpty RopeNodeKind = iota
	RopeKindParent
	RopeKindLeaf
)

type RopeNode interface {
	Kind() RopeNodeKind
	Length() int
	String() string
	Runes() []rune
	Rebalance() RopeNode
	Split(offset int) (left RopeNode, right RopeNode)
	Insert(offset int, text string) RopeNode
	Delete(offset int, count int) RopeNode
	GetLineNumber(offset int) int
	GetLineStart(line int) int
	GetLineEnd(line int) int
	PrettyPrint(w io.Writer, indent string)
}

type RopeEmptyNode struct{}

type RopeParentNode struct {
	length     int
	LineBreaks int
	left       RopeNode
	right      RopeNode
}

type RopeLeafNode struct {
	LineBreaks int
	text       []rune
}

func (n *RopeEmptyNode) Kind() RopeNodeKind {
	return RopeKindEmpty
}

func (n *RopeParentNode) Kind() RopeNodeKind {
	return RopeKindParent
}

func (n *RopeLeafNode) Kind() RopeNodeKind {
	return RopeKindLeaf
}

func New(text string) RopeNode {
	if text == "" {
		return &RopeEmptyNode{}
	} else {
		runes := []rune(text)

		return &RopeLeafNode{countLF(runes), runes}
	}
}
