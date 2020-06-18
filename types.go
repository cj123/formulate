package formulate

import (
	"golang.org/x/net/html"
	"net/url"
	"reflect"
)

type (
	Password string
	Email    string
	URL      string
	Tel      string
)

type Select interface {
	// SelectMultiple indicates whether multiple options can be selected at once.
	SelectMultiple() bool

	// SelectOptions are the available options
	SelectOptions() []Option
}

type Option struct {
	Value interface{}
	Label string

	Disabled bool
	Checked  *Condition
	Attr     []html.Attribute
}

type Condition bool

func NewCondition(b bool) *Condition {
	c := Condition(b)

	return &c
}

type Radio interface {
	RadioOptions() []Option
}

type CustomEncoder interface {
	BuildFormElement(key string, parent *html.Node, field StructField, decorator Decorator)
}

type CustomDecoder interface {
	DecodeFormValue(form url.Values, name string, values []string) (reflect.Value, error)
}

type BoolNumber int

func (bn BoolNumber) DecodeFormValue(form url.Values, name string, values []string) (reflect.Value, error) {
	val := values[0]

	if val == "on" || val == "1" {
		return reflect.ValueOf(BoolNumber(1)), nil
	}

	return reflect.ValueOf(BoolNumber(0)), nil
}

func (bn *BoolNumber) Bool() bool {
	return *bn == 1
}

type Raw []byte

func (r Raw) BuildFormElement(key string, parent *html.Node, field StructField, decorator Decorator) {
	n := &html.Node{
		Type: html.ElementNode,
		Data: "textarea",
		Attr: []html.Attribute{
			{
				Key: "name",
				Val: key,
			},
			{
				Key: "id",
				Val: key,
			},
		},
	}

	n.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: string(r),
	})

	parent.AppendChild(n)
	decorator.TextareaField(n, field)
}
