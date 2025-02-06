package rope

import (
	"encoding/json"
	"io"
	"strconv"
)

func (n *RopeEmptyNode) PrettyPrint(w io.Writer, indent string) {
	w.Write([]byte("<Empty>"))
}

func (n *RopeLeafNode) PrettyPrint(w io.Writer, indent string) {
	text, _ := json.Marshal(string(n.text))
	br := strconv.Itoa(n.LineBreaks)

	w.Write([]byte("<Leaf: breaks: " + br + " text: " + string(text) + ">"))
}

func (n *RopeParentNode) PrettyPrint(w io.Writer, indent string) {
	w.Write([]byte("<Parent:\n" + indent + "left:\n" + indent + "  "))

	n.left.PrettyPrint(w, indent+"  ")

	w.Write([]byte("\n" + indent + "right:\n" + indent + "  "))

	n.right.PrettyPrint(w, indent+"  ")

	w.Write([]byte("\n>"))
}
