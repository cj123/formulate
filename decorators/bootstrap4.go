package decorators

import (
	"github.com/cj123/formulate"
	"golang.org/x/net/html"
)

// BootstrapDecorator implements a form layout using Bootstrap 4.
type BootstrapDecorator struct{}

var _ formulate.Decorator = &BootstrapDecorator{}

func (b BootstrapDecorator) FieldWrapper(n *html.Node, field formulate.StructField) {
	b.col8(n)
}

func (b BootstrapDecorator) HelpText(n *html.Node, field formulate.StructField) {
	n.Data = "div"
	formulate.AppendClass(n, "small mt-1")
}

func (b BootstrapDecorator) RootNode(n *html.Node) {

}

func (b BootstrapDecorator) RadioButton(n *html.Node, field formulate.StructField) {

}

func (b BootstrapDecorator) Fieldset(n *html.Node, field formulate.StructField) {

}

func (b BootstrapDecorator) Row(n *html.Node, field formulate.StructField) {
	formulate.AppendClass(n, "row", "form-group")
}

func (b BootstrapDecorator) TextField(n *html.Node, field formulate.StructField) {
	b.formControl(n)
}

func (b BootstrapDecorator) Label(n *html.Node, field formulate.StructField) {
	b.col4(n)
}

func (b BootstrapDecorator) col4(n *html.Node) {
	formulate.AppendClass(n, "col-md-4 col-12")
}

func (b BootstrapDecorator) col8(n *html.Node) {
	formulate.AppendClass(n, "col-md-8 col-12")
}

func (b BootstrapDecorator) formControl(n *html.Node) {
	formulate.AppendClass(n, "form-control")
}

func (b BootstrapDecorator) NumberField(n *html.Node, field formulate.StructField) {
	b.formControl(n)
}

func (b BootstrapDecorator) CheckboxField(n *html.Node, field formulate.StructField) {
}

func (b BootstrapDecorator) TextareaField(n *html.Node, field formulate.StructField) {
	b.formControl(n)
}

func (b BootstrapDecorator) TimeField(n *html.Node, field formulate.StructField) {
	b.formControl(n)
}

func (b BootstrapDecorator) SelectField(n *html.Node, field formulate.StructField) {
	b.formControl(n)
}
