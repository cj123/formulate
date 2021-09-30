package formulate

import "golang.org/x/net/html"

// Decorator is used to customise node elements that are built by the HTMLEncoder.
// Custom decorators can be passed into the HTMLEncoder.
// If no decorator is specified, a nil decorator is used. This applies no decoration to the output HTML.
type Decorator interface {
	// RootNode decorates the root <div> of the returned HTML.
	RootNode(n *html.Node)
	// Fieldset decorates each <fieldset>. Fieldsets are created for each
	// non-anonymous struct within the encoded data structure.
	Fieldset(n *html.Node, field StructField)
	// Row decorates the parent of each label, input and help text, for each field within the encoded data structure.
	Row(n *html.Node, field StructField)
	// FieldWrapper decorates the div which wraps the input and help text within a form
	FieldWrapper(n *html.Node, field StructField)
	// Label decorates the <label> for the form element
	Label(n *html.Node, field StructField)
	// HelpText decorates the text which is displayed below each form element.
	// The HelpText is generated from the "help" struct tag.
	HelpText(n *html.Node, field StructField)
	// TextField decorates an <input type="text">
	TextField(n *html.Node, field StructField)
	// NumberField decorates an <input type="number"> or equivalent (e.g. Tel)
	NumberField(n *html.Node, field StructField)
	// CheckboxField decorates an <input type="checkbox">, displayed for boolean values.
	CheckboxField(n *html.Node, field StructField)
	// TextareaField decorates a <textarea> tag.
	TextareaField(n *html.Node, field StructField)
	// TimeField decorates an <input type="datetime-local"> used to represent time values.
	TimeField(n *html.Node, field StructField)
	// SelectField decorates a <select> dropdown
	SelectField(n *html.Node, field StructField)
	// RadioButton decorates an individual <input type="radio">
	RadioButton(n *html.Node, field StructField)
	// ValidationText decorates the text which is displayed below each form element when there is a validation error.
	ValidationText(n *html.Node, field StructField)
}

type nilDecorator struct{}

func (d nilDecorator) RootNode(n *html.Node) {}

func (d nilDecorator) Fieldset(n *html.Node, field StructField) {}

func (d nilDecorator) Row(n *html.Node, field StructField) {}

func (d nilDecorator) FieldWrapper(n *html.Node, field StructField) {}

func (d nilDecorator) Label(n *html.Node, field StructField) {}

func (d nilDecorator) HelpText(n *html.Node, field StructField) {}

func (d nilDecorator) TextField(n *html.Node, field StructField) {}

func (d nilDecorator) NumberField(n *html.Node, field StructField) {}

func (d nilDecorator) CheckboxField(n *html.Node, field StructField) {}

func (d nilDecorator) TextareaField(n *html.Node, field StructField) {}

func (d nilDecorator) TimeField(n *html.Node, field StructField) {}

func (d nilDecorator) SelectField(n *html.Node, field StructField) {}

func (d nilDecorator) RadioButton(n *html.Node, field StructField) {}

func (d nilDecorator) ValidationText(n *html.Node, field StructField) {}
