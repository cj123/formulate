package formulate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/yosssi/gohtml"
	"golang.org/x/net/html"
)

// HTMLEncoder is used to generate a HTML form from a given struct.
type HTMLEncoder struct {
	n *html.Node
	w io.Writer

	decorator Decorator
	format    bool

	showConditions map[string]ShowConditionFunc
}

// NewEncoder returns a HTMLEncoder which outputs to w. A Decorator can be passed to NewEncoder, which will then be used
// to style the outputted HTML. If nil is passed in, no decorator is used, and a barebones HTML form will be returned.
func NewEncoder(w io.Writer, decorator Decorator) *HTMLEncoder {
	n := &html.Node{
		Type: html.ElementNode,
		Data: "div",
	}

	if decorator == nil {
		decorator = nilDecorator{}
	}

	decorator.RootNode(n)

	return &HTMLEncoder{
		w:              w,
		n:              n,
		decorator:      decorator,
		showConditions: make(map[string]ShowConditionFunc),
	}
}

// SetFormat tells the HTMLEncoder to output formatted HTML.
// Formatting is provided by the https://github.com/yosssi/gohtml package.
func (h *HTMLEncoder) SetFormat(b bool) {
	h.format = b
}

// ShowConditionFunc is a function which determines whether or not to show a form element. See: HTMLEncoder.AddShowCondition
type ShowConditionFunc func() bool

// AddShowCondition allows you to determine visibility of certain form elements.
// For example, given the following struct:
//   type Example struct {
//     Name string
//     SecretOption bool `show:"adminOnly"`
//   }
// If you wanted to make the SecretOption field only show to admins, you would call AddShowCondition as follows:
//   AddShowCondition("adminOnly", func() bool {
//      // some code that determines if we are 'admin'
//   })
// You can add multiple ShowConditions, but they must have different keys.
func (h *HTMLEncoder) AddShowCondition(key string, fn ShowConditionFunc) {
	h.showConditions[key] = fn
}

func errorIncorrectValue(t reflect.Type) error {
	return fmt.Errorf("formulate: encode expects a struct value, got: %s", t.String())
}

// Encode takes a struct (or struct pointer) and produces a HTML form from all elements in the struct.
// The encoder deals with most simple types and structs, but more complex types (maps, slices, arrays)
// will render as a JSON blob in a <textarea>.
//
// The rendering behavior of any element can be replaced by implementing the CustomEncoder interface.
func (h *HTMLEncoder) Encode(i interface{}) error {
	v := reflect.ValueOf(i)

	if v.Kind() == reflect.Ptr {
		if !v.IsValid() || v.Elem().Kind() != reflect.Struct {
			return errorIncorrectValue(v.Type())
		}
	} else if v.Kind() != reflect.Struct {
		return errorIncorrectValue(v.Type())
	}

	if err := h.recurse(v, v.Type().String(), StructField{}, h.n); err != nil {
		return err
	}

	if !h.format {
		return html.Render(h.w, h.n)
	}

	buf := new(bytes.Buffer)

	if err := html.Render(buf, h.n); err != nil {
		return err
	}

	if _, err := h.w.Write(gohtml.FormatBytes(buf.Bytes())); err != nil {
		return err
	}

	return nil
}

func (h *HTMLEncoder) recurse(v reflect.Value, key string, field StructField, parent *html.Node) error {
	switch v.Interface().(type) {
	case time.Time, Select, RadioList, CustomEncoder:
		return BuildField(v, formElementName(key), field, parent, h.decorator, h.showConditions)
	}

	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() && v.CanAddr() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		return h.recurse(v.Elem(), key, field, parent)
	case reflect.Struct:
		if field.Hidden(h.showConditions) {
			return nil
		}

		if field.BuildFieldset() {
			// anonymous structs use their parent's fieldset
			parent = h.buildFieldSet(field, parent)
		}

		for i := 0; i < v.NumField(); i++ {
			structField := v.Type().Field(i)

			err := h.recurse(v.Field(i), key+fieldSeparator+v.Type().Field(i).Name, StructField{structField}, parent)

			if err != nil {
				return err
			}
		}
		return nil
	case reflect.Slice, reflect.Array, reflect.Map:
		buf := new(bytes.Buffer)

		enc := json.NewEncoder(buf)
		enc.SetIndent("", "  ")

		if err := enc.Encode(v.Interface()); err != nil {
			return err
		}

		return h.recurse(reflect.ValueOf(Raw(buf.Bytes())), key, field, parent)
	default:
		return BuildField(v, formElementName(key), field, parent, h.decorator, h.showConditions)
	}
}

func (h *HTMLEncoder) buildFieldSet(field StructField, parent *html.Node) *html.Node {
	n := &html.Node{
		Type: html.ElementNode,
		Data: "fieldset",
	}

	name := field.GetName()

	if name != "" {
		legend := &html.Node{
			Type: html.ElementNode,
			Data: "legend",
		}

		legend.AppendChild(&html.Node{
			Type: html.TextNode,
			Data: name,
		})

		n.AppendChild(legend)
	}

	parent.AppendChild(n)
	h.decorator.Fieldset(n, field)

	return n
}

func BuildField(v reflect.Value, key string, field StructField, parent *html.Node, decorator Decorator, showConditions map[string]ShowConditionFunc) error {
	if !v.IsValid() || field.Hidden(showConditions) {
		return nil
	}

	rowElement := &html.Node{
		Type: html.ElementNode,
		Data: "div",
	}

	BuildLabel(key, rowElement, field, decorator)
	wrapper := &html.Node{
		Type: html.ElementNode,
		Data: "div",
	}

	rowElement.AppendChild(wrapper)
	decorator.FieldWrapper(wrapper, field)

	parent.AppendChild(rowElement)

	defer func() {
		BuildHelpText(wrapper, field, decorator)
		decorator.Row(rowElement, field)
	}()

	switch a := v.Interface().(type) {
	case CustomEncoder:
		return a.BuildFormElement(key, wrapper, field, decorator)
	case time.Time:
		n := BuildTimeField(a, key, field)
		wrapper.AppendChild(n)
		decorator.NumberField(n, field)
		return nil
	case Select:
		n := BuildSelectField(a, key, field)
		wrapper.AppendChild(n)
		decorator.SelectField(n, field)
		return nil
	case RadioList:
		n := BuildRadioButtons(a, key, field, decorator)
		wrapper.AppendChild(n)
		return nil
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float64, reflect.Float32:
		if _, ok := v.Interface().(BoolNumber); ok {
			n := BuildBoolField(v, key, field)
			wrapper.AppendChild(n)
			decorator.CheckboxField(n, field)
		} else {
			n := BuildNumberField(v, key, field)
			wrapper.AppendChild(n)
			decorator.NumberField(n, field)
		}
		return nil
	case reflect.String:
		n := BuildStringField(v, key, field)
		wrapper.AppendChild(n)
		decorator.TextareaField(n, field)
		return nil
	case reflect.Bool:
		n := BuildBoolField(v, key, field)
		wrapper.AppendChild(n)
		decorator.CheckboxField(n, field)
		return nil
	default:
		panic("formulate: unknown element kind: " + v.Kind().String())
	}
}

const timeFormat = "2006-01-02T15:04"

func BuildTimeField(t time.Time, key string, field StructField) *html.Node {
	n := &html.Node{
		Type: html.ElementNode,
		Data: "input",
		Attr: []html.Attribute{
			{
				Key: "type",
				Val: "datetime-local", // @TODO consider replacing use of datetime-local with a time and date input
			},
			{
				Key: "name",
				Val: key,
			},
			{
				Key: "id",
				Val: key,
			},
			{
				Key: "value",
				Val: t.Format(timeFormat),
			},
		},
	}

	if field.HasMin() {
		n.Attr = append(n.Attr, html.Attribute{
			Key: "min",
			Val: field.Min(),
		})
	}

	if field.HasMax() {
		n.Attr = append(n.Attr, html.Attribute{
			Key: "max",
			Val: field.Max(),
		})
	}

	return n
}

func BuildNumberField(v reflect.Value, key string, field StructField) *html.Node {
	n := &html.Node{
		Type: html.ElementNode,
		Data: "input",
		Attr: []html.Attribute{
			{
				Key: "type",
				Val: "number",
			},
			{
				Key: "name",
				Val: key,
			},
			{
				Key: "id",
				Val: key,
			},
			{
				Key: "value",
				Val: toString(v.Interface()),
			},
		},
	}

	if field.HasMin() {
		n.Attr = append(n.Attr, html.Attribute{
			Key: "min",
			Val: field.Min(),
		})
	}

	if field.HasMax() {
		n.Attr = append(n.Attr, html.Attribute{
			Key: "max",
			Val: field.Max(),
		})
	}

	if field.HasStep() {
		n.Attr = append(n.Attr, html.Attribute{
			Key: "step",
			Val: field.Step(),
		})
	} else if v.Kind() == reflect.Float64 || v.Kind() == reflect.Float32 {
		n.Attr = append(n.Attr, html.Attribute{
			Key: "step",
			Val: "any",
		})
	}

	return n
}

func BuildStringField(v reflect.Value, key string, field StructField) *html.Node {
	var n *html.Node

	if field.Elem() == "textarea" {
		n = &html.Node{
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
			Data: v.String(),
		})
	} else {
		typField := func() string {
			switch v.Interface().(type) {
			case Password:
				return "password"
			case Email:
				return "email"
			case URL:
				return "url"
			case Tel:
				return "tel"
			default:
				return "text"
			}
		}

		n = &html.Node{
			Type: html.ElementNode,
			Data: "input",
			Attr: []html.Attribute{
				{
					Key: "type",
					Val: field.InputType(typField()),
				},
				{
					Key: "name",
					Val: key,
				},
				{
					Key: "id",
					Val: key,
				},
				{
					Key: "value",
					Val: v.String(),
				},
			},
		}

		if pattern := field.Pattern(); pattern != "" {
			n.Attr = append(n.Attr, html.Attribute{
				Key: "pattern",
				Val: pattern,
			})
		}
	}

	if placeholder := field.Placeholder(); placeholder != "" {
		n.Attr = append(n.Attr, html.Attribute{
			Key: "placeholder",
			Val: placeholder,
		})
	}

	if field.Required() {
		n.Attr = append(n.Attr, html.Attribute{
			Key: "required",
			Val: "required",
		})
	}

	return n
}

func BuildBoolField(v reflect.Value, key string, field StructField) *html.Node {
	n := &html.Node{
		Type: html.ElementNode,
		Data: "input",
		Attr: []html.Attribute{
			{
				Key: "type",
				Val: "checkbox",
			},
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

	checked := false

	if bn, ok := v.Interface().(BoolNumber); ok {
		if bn.Bool() {
			checked = true
		}
	} else if v.Bool() {
		checked = true
	}

	if checked {
		n.Attr = append(n.Attr, html.Attribute{Key: "checked", Val: "checked"})
	}

	return n
}

func BuildSelectField(s Select, key string, field StructField) *html.Node {
	sel := &html.Node{
		Type: html.ElementNode,
		Data: "select",
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

	if s.SelectMultiple() {
		sel.Attr = append(sel.Attr, html.Attribute{
			Key: "multiple",
		})
	}

	optGroups := make(map[string]*html.Node)

	selectOptions := s.SelectOptions()

	for _, opt := range selectOptions {
		if opt.Group == nil {
			continue
		}

		if _, ok := optGroups[*opt.Group]; ok {
			continue
		}

		optGroups[*opt.Group] = &html.Node{
			Type: html.ElementNode,
			Data: "optgroup",
			Attr: []html.Attribute{
				{
					Key: "label",
					Val: *opt.Group,
				},
			},
		}
	}

	for _, opt := range selectOptions {
		o := &html.Node{
			Type: html.ElementNode,
			Data: "option",
			Attr: []html.Attribute{
				{
					Key: "value",
					Val: toString(opt.Value),
				},
			},
		}

		if opt.Disabled {
			o.Attr = append(o.Attr, html.Attribute{Key: "disabled"})
		}

		checked := false

		if opt.Checked == nil {
			checked = opt.Value == s
		} else {
			checked = bool(*opt.Checked)
		}

		if checked {
			o.Attr = append(o.Attr, html.Attribute{Key: "selected"})
		}

		o.Attr = append(o.Attr, opt.Attr...)

		o.AppendChild(&html.Node{
			Type: html.TextNode,
			Data: opt.Label,
		})

		if opt.Group != nil {
			optGroups[*opt.Group].AppendChild(o)
		} else {
			sel.AppendChild(o)
		}
	}

	addedOptGroups := make(map[string]bool)

	for _, opt := range selectOptions {
		if opt.Group == nil {
			continue
		}

		if _, ok := addedOptGroups[*opt.Group]; ok {
			continue
		}

		sel.AppendChild(optGroups[*opt.Group])
		addedOptGroups[*opt.Group] = true
	}

	return sel
}

func BuildRadioButtons(r RadioList, key string, field StructField, decorator Decorator) *html.Node {
	div := &html.Node{
		Type: html.ElementNode,
		Data: "div",
		Attr: []html.Attribute{
			{
				Key: "id",
				Val: key,
			},
		},
	}

	for i, opt := range r.RadioOptions() {
		id := fmt.Sprintf("%s%d", key, i)

		radio := &html.Node{
			Type: html.ElementNode,
			Data: "input",
			Attr: []html.Attribute{
				{
					Key: "type",
					Val: "radio",
				},
				{
					Key: "value",
					Val: toString(opt.Value),
				},
				{
					Key: "id",
					Val: id,
				},
				{
					Key: "name",
					Val: key,
				},
			},
		}

		if opt.Disabled {
			radio.Attr = append(radio.Attr, html.Attribute{Key: "disabled"})
		}

		radio.Attr = append(radio.Attr, opt.Attr...)

		checked := false

		if opt.Checked == nil {
			checked = opt.Value == r
		} else {
			checked = bool(*opt.Checked)
		}

		if checked {
			radio.Attr = append(radio.Attr, html.Attribute{Key: "checked"})
		}

		label := &html.Node{
			Type: html.ElementNode,
			Data: "label",
			Attr: []html.Attribute{
				{
					Key: "for",
					Val: id,
				},
			},
		}

		label.AppendChild(&html.Node{
			Type: html.TextNode,
			Data: opt.Label,
		})

		div.AppendChild(label)
		div.AppendChild(radio)

		decorator.Label(label, field)
		decorator.RadioButton(radio, field)
	}

	return div
}

func formElementName(label string) string {
	return strings.Join(strings.Split(label, fieldSeparator)[2:], fieldSeparator)
}

func BuildLabel(label string, parent *html.Node, field StructField, decorator Decorator) {
	n := &html.Node{
		Type: html.ElementNode,
		Data: "label",
		Attr: []html.Attribute{
			{
				Key: "for",
				Val: label,
			},
		},
	}

	n.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: field.GetName(),
	})

	parent.AppendChild(n)
	decorator.Label(n, field)
}

func BuildHelpText(parent *html.Node, field StructField, decorator Decorator) {
	helpText := field.GetHelpText()

	n := &html.Node{
		Type: html.ElementNode,
		Data: "div",
	}

	n.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: helpText,
	})

	parent.AppendChild(n)
	decorator.HelpText(n, field)
}

func toString(i interface{}) string {
	val := reflect.ValueOf(i)

	switch val.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(val.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(val.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(val.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(val.Float(), 'f', -1, 64)
	case reflect.String:
		return val.String()
	default:
		return fmt.Sprintf("%v", i)
	}
}
