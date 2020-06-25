package formulate

import (
	"net/url"
	"reflect"

	"golang.org/x/net/html"
)

// CustomEncoder allows for custom rendering behavior of a type to be implemented.
// If a type implements the CustomEncoder interface, BuildFormElement is called in place
// of any other formulate rendering behavior for inputs. The label and help text of the element
// will still be rendered within the row as normal.
type CustomEncoder interface {
	// BuildFormElement is passed the key of the form element as computed by formulate,
	// the parent node of the element, the field of the struct
	// that is currently being rendered, and the form's decorator.
	// Note that the built element must be appended to the parent or it will not be shown in the form!
	// Errors returned from BuildFormElement propagate back through to the formulate.Encoder.Encode call.
	BuildFormElement(key string, parent *html.Node, field StructField, decorator Decorator) error
}

// CustomDecoder allows for custom decoding behavior to be specified for an element. If
// a type implements the CustomDecoder interface, DecodeFormValue is called in place of
// any other decoding behavior.
type CustomDecoder interface {
	// DecodeFormValue is passed the whole form values, the name of the element that it is decoding,
	// and the values for that specific element. It must return a reflect.Value of equal type to the
	// type which is implementing the CustomDecoder interface. If err != nil, the error will propagate
	// back through to the Decode() call.
	DecodeFormValue(form url.Values, name string, values []string) (reflect.Value, error)
}

type (
	// Password represents an <input type="password">
	Password string
	// Email represents an <input type="email">
	Email string
	// URL represents an <input type="url">
	URL string
	// Tel represents an <input type="tel">
	Tel string
)

// Select represents a HTML <select> element.
type Select interface {
	// SelectMultiple indicates whether multiple options can be selected at once.
	SelectMultiple() bool

	// SelectOptions are the available options
	SelectOptions() []Option
}

// Option represents an option in Select inputs and Radio inputs.
type Option struct {
	Value interface{}
	Label string

	Disabled bool
	Checked  *Condition
	Attr     []html.Attribute
}

// Condition are optional booleans for Options.
type Condition bool

// NewCondition creates a new condition based on a bool value.
func NewCondition(b bool) *Condition {
	c := Condition(b)

	return &c
}

// RadioList represents a list of <input type="radio">. Radio lists must implement their own decoder.
type RadioList interface {
	CustomDecoder

	RadioOptions() []Option
}

// BoolNumber represents an int (0 or 1) which should actually be rendered as a checkbox.
// It is provided here as a convenience, as many structures use 0 or 1 to represent booleans values.
type BoolNumber int

// DecodeFormValue implements the CustomDecoder interface.
func (bn BoolNumber) DecodeFormValue(form url.Values, name string, values []string) (reflect.Value, error) {
	val := values[0]

	if val == "on" || val == "1" {
		return reflect.ValueOf(BoolNumber(1)), nil
	}

	return reflect.ValueOf(BoolNumber(0)), nil
}

// Bool returns true if the underlying value is 1.
func (bn *BoolNumber) Bool() bool {
	return *bn == 1
}

// Raw is byte data which should be rendered as a string inside a textarea.
type Raw []byte

// BuildFormElement implements the CustomEncoder interface.
func (r Raw) BuildFormElement(key string, parent *html.Node, field StructField, decorator Decorator) error {
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

	return nil
}
