package formulate

import "golang.org/x/net/html"

type Decorator interface {
	Form(n *html.Node)
	Fieldset(n *html.Node, field StructField)
	Row(n *html.Node, field StructField)
	FieldWrapper(n *html.Node, field StructField)
	Label(n *html.Node, field StructField)
	HelpText(n *html.Node, field StructField)

	TextField(n *html.Node, field StructField)
	NumberField(n *html.Node, field StructField)
	CheckboxField(n *html.Node, field StructField)
	TextareaField(n *html.Node, field StructField)
	TimeField(n *html.Node, field StructField)
	SelectField(n *html.Node, field StructField)
	RadioButton(n *html.Node, field StructField)
}

type nilDecorator struct{}

func (d nilDecorator) FieldWrapper(n *html.Node, field StructField) {

}

func (d nilDecorator) HelpText(n *html.Node, field StructField) {

}

func (d nilDecorator) Form(n *html.Node) {
}

func (d nilDecorator) RadioButton(n *html.Node, field StructField) {

}

func (d nilDecorator) SelectField(n *html.Node, field StructField) {

}

func (d nilDecorator) Fieldset(n *html.Node, field StructField) {

}

func (d nilDecorator) TimeField(n *html.Node, field StructField) {

}

func (d nilDecorator) Row(n *html.Node, field StructField) {

}

func (d nilDecorator) TextField(n *html.Node, field StructField) {

}

func (d nilDecorator) Label(n *html.Node, field StructField) {

}

func (d nilDecorator) NumberField(n *html.Node, field StructField) {

}

func (d nilDecorator) CheckboxField(n *html.Node, field StructField) {

}

func (d nilDecorator) TextareaField(n *html.Node, field StructField) {

}

type BootstrapDecorator struct{}

func (b BootstrapDecorator) FieldWrapper(n *html.Node, field StructField) {
	b.col8(n)
}

func (b BootstrapDecorator) HelpText(n *html.Node, field StructField) {
	n.Data = "div"
	AppendClass(n, "small mt-1")
}

func (b BootstrapDecorator) Form(n *html.Node) {

}

func (b BootstrapDecorator) RadioButton(n *html.Node, field StructField) {

}

func (b BootstrapDecorator) Fieldset(n *html.Node, field StructField) {

}

func (b BootstrapDecorator) Row(n *html.Node, field StructField) {
	AppendClass(n, "row", "form-group")
}

func (b BootstrapDecorator) TextField(n *html.Node, field StructField) {
	b.formControl(n)
}

func (b BootstrapDecorator) Label(n *html.Node, field StructField) {
	b.col4(n)
}

func (b BootstrapDecorator) col4(n *html.Node) {
	AppendClass(n, "col-md-4 col-12")
}

func (b BootstrapDecorator) col8(n *html.Node) {
	AppendClass(n, "col-md-8 col-12")
}

func (b BootstrapDecorator) offset4(n *html.Node) {
	AppendClass(n, "offset-md-4")
}

func (b BootstrapDecorator) formControl(n *html.Node) {
	AppendClass(n, "form-control")
}

func (b BootstrapDecorator) NumberField(n *html.Node, field StructField) {
	b.formControl(n)
}

func (b BootstrapDecorator) CheckboxField(n *html.Node, field StructField) {
}

func (b BootstrapDecorator) TextareaField(n *html.Node, field StructField) {
	b.formControl(n)
}

func (b BootstrapDecorator) TimeField(n *html.Node, field StructField) {
	b.formControl(n)
}

func (b BootstrapDecorator) SelectField(n *html.Node, field StructField) {
	b.formControl(n)
}
