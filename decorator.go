package formulate

import "golang.org/x/net/html"

type Decorator interface {
	Fieldset(n *html.Node)
	Row(n *html.Node)
	TextField(n *html.Node)
	Label(n *html.Node)
	NumberField(n *html.Node)
	CheckboxField(n *html.Node)
	TextareaField(n *html.Node)
	TimeField(n *html.Node)
	SelectField(n *html.Node)
	RadioButton(n *html.Node)
}

type nilDecorator struct{}

func (d nilDecorator) RadioButton(n *html.Node) {

}

func (d nilDecorator) SelectField(n *html.Node) {

}

func (d nilDecorator) Fieldset(n *html.Node) {

}

func (d nilDecorator) TimeField(n *html.Node) {

}

func (d nilDecorator) Row(n *html.Node) {

}

func (d nilDecorator) TextField(n *html.Node) {

}

func (d nilDecorator) Label(n *html.Node) {

}

func (d nilDecorator) NumberField(n *html.Node) {

}

func (d nilDecorator) CheckboxField(n *html.Node) {

}

func (d nilDecorator) TextareaField(n *html.Node) {

}

type bootstrapDecorator struct{}

func (b bootstrapDecorator) RadioButton(n *html.Node) {

}

func (b bootstrapDecorator) Fieldset(n *html.Node) {

}

func (b bootstrapDecorator) Row(n *html.Node) {
	AppendClass(n, "row", "form-group")
}

func (b bootstrapDecorator) TextField(n *html.Node) {
	b.col8(n)
	b.formControl(n)
}

func (b bootstrapDecorator) Label(n *html.Node) {
	b.col4(n)
}

func (b bootstrapDecorator) col4(n *html.Node) {
	AppendClass(n, "col-md-4 col-12")
}

func (b bootstrapDecorator) col8(n *html.Node) {
	AppendClass(n, "col-md-8 col-12")
}

func (b bootstrapDecorator) formControl(n *html.Node) {
	AppendClass(n, "form-control")
}

func (b bootstrapDecorator) NumberField(n *html.Node) {
	b.col8(n)
	b.formControl(n)
}

func (b bootstrapDecorator) CheckboxField(n *html.Node) {
	b.col8(n)
}

func (b bootstrapDecorator) TextareaField(n *html.Node) {
	b.col8(n)
	b.formControl(n)
}

func (b bootstrapDecorator) TimeField(n *html.Node) {
	b.col8(n)
	b.formControl(n)
}

func (b bootstrapDecorator) SelectField(n *html.Node) {
	b.col8(n)
	b.formControl(n)
}
