package formulate

import (
	"html/template"
	"strings"

	"golang.org/x/net/html"
)

// RenderHTMLToNode renders a html string into a parent *html.Node.
// The parent node is emptied before the data is rendered into it.
func RenderHTMLToNode(data template.HTML, parent *html.Node) error {
	n, err := html.Parse(strings.NewReader(string(data)))

	if err != nil {
		return err
	}

	body := getBody(n)

	for c := parent.FirstChild; c != nil; c = c.NextSibling {
		parent.RemoveChild(c)
	}

	moveNodeChildren(body, parent)

	return nil
}

// getBody finds the <body> element inside the parsed HTML tree.
func getBody(node *html.Node) *html.Node {
	var body *html.Node

	var f func(n *html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "body" {
			body = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(node)

	return body
}

func moveNodeChildren(from, to *html.Node) {
	c := from.FirstChild

	for c != nil {
		x := c
		c = c.NextSibling

		from.RemoveChild(x)
		to.AppendChild(x)
	}
}
