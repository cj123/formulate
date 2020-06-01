package formulate

import (
	"golang.org/x/net/html"
	"net/url"
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
	BuildFormElement(key string, parent *html.Node, field StructField)
}

type CustomDecoder interface {
	DecodeFormValue(form url.Values, name string, values []string) error
}

type BoolNumber int

func (bn BoolNumber) Bool() bool {
	return bn == 1
}
