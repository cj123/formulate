package formulate

import (
	"strings"

	"golang.org/x/net/html"
)

// AppendClass adds a class to a HTML node.
func AppendClass(n *html.Node, classes ...string) {
	class := strings.Join(classes, " ")

	for i, attr := range n.Attr {
		if attr.Key == "class" {
			n.Attr[i].Val += " " + class
			return
		}
	}

	n.Attr = append(n.Attr, html.Attribute{
		Key: "class",
		Val: class,
	})
}

// HasAttribute returns true if n has the attribute named attr.
func HasAttribute(n *html.Node, attr string) bool {
	for _, a := range n.Attr {
		if a.Key == attr {
			return true
		}
	}

	return false
}
