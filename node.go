package formulate

import (
	"strings"

	"golang.org/x/net/html"
)

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

func HasAttribute(n *html.Node, attr string) bool {
	for _, a := range n.Attr {
		if a.Key == attr {
			return true
		}
	}

	return false
}
