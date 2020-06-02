package formulate

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type Encoder interface {
	Encode(i interface{}) error
	AddShowCondition(value string, fn ShowConditionFunc)
}

type ShowConditionFunc func() bool

type htmlEncoder struct {
	n *html.Node
	w io.Writer

	decorator Decorator

	showConditions map[string]ShowConditionFunc
}

func NewEncoder(w io.Writer, decorator Decorator) Encoder {
	n := &html.Node{
		Type: html.ElementNode,
		Data: "div",
	}

	if decorator == nil {
		decorator = nilDecorator{}
	}

	decorator.Form(n)

	return &htmlEncoder{
		w:              w,
		n:              n,
		decorator:      decorator,
		showConditions: make(map[string]ShowConditionFunc),
	}
}

func (h *htmlEncoder) AddShowCondition(value string, fn ShowConditionFunc) {
	h.showConditions[value] = fn
}

func (h *htmlEncoder) Encode(i interface{}) error {
	v := reflect.ValueOf(i)

	h.recurse(v, v.Type().String(), StructField{}, h.n)

	return html.Render(h.w, h.n)
}

func (h *htmlEncoder) recurse(v reflect.Value, key string, field StructField, parent *html.Node) {
	switch v.Interface().(type) {
	case time.Time, Select, Radio, CustomEncoder:
		h.buildField(v, key, field, parent)
		return
	}

	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() && v.CanAddr() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		h.recurse(v.Elem(), key, field, parent)
		return
	case reflect.Struct:
		if !field.Anonymous {
			// anonymous structs use their parent's fieldset
			parent = h.buildFieldSet(v, field, parent)
		}

		for i := 0; i < v.NumField(); i++ {
			structField := v.Type().Field(i)

			h.recurse(v.Field(i), key+"."+v.Type().Field(i).Name, StructField{structField}, parent)
		}
		return
	case reflect.Map:
		iter := v.MapRange()

		for iter.Next() {
			// something
			h.recurse(iter.Value(), key, field, parent)
		}

		// @TODO controls to add/remove a map value?
		return
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			val := v.Index(i)

			h.recurse(val, key, field, parent)
		}

		// @TODO controls to add/remove an array slice?
		return
	default:
		h.buildField(v, key, field, parent)
		return
	}
}

func (h *htmlEncoder) buildFieldSet(v reflect.Value, field StructField, parent *html.Node) *html.Node {
	n := &html.Node{
		Type: html.ElementNode,
		Data: "fieldset",
	}

	legend := &html.Node{
		Type: html.ElementNode,
		Data: "legend",
	}

	name := field.GetName()

	if name == "" {
		name = camelCase(v.Type().Name())
	}

	legend.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: name,
	})

	n.AppendChild(legend)

	h.decorator.Fieldset(n)

	parent.AppendChild(n)

	return n
}

func (h *htmlEncoder) buildField(v reflect.Value, key string, field StructField, parent *html.Node) {
	if !v.IsValid() || field.Hidden(h.showConditions) {
		return
	}

	rowElement := &html.Node{
		Type: html.ElementNode,
		Data: "div",
	}

	h.decorator.Row(rowElement)

	h.buildLabel(key, rowElement, field)

	defer func() {
		h.buildHelpText(key, rowElement, field)
		parent.AppendChild(rowElement)
	}()

	key = h.formElementName(key)

	switch a := v.Interface().(type) {
	case CustomEncoder:
		a.BuildFormElement(key, rowElement, field)
		return
	case time.Time:
		h.buildTimeField(a, key, rowElement, field)
		return
	case Select:
		h.buildSelectField(a, key, rowElement, field)
		return
	case Radio:
		h.buildRadioButtons(a, key, rowElement, field)
		return
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float64, reflect.Float32:
		if _, ok := v.Interface().(BoolNumber); ok {
			h.buildBoolField(v, key, rowElement)
		} else {
			h.buildNumberField(v, key, rowElement, field)
		}
		return
	case reflect.String:
		h.buildStringField(v, key, rowElement, field)
		return
	case reflect.Bool:
		h.buildBoolField(v, key, rowElement)
		return
	}

	panic("formulate: unknown element kind: " + v.Kind().String())
}

const timeFormat = "2006-01-02T15:04"

func (h *htmlEncoder) buildTimeField(t time.Time, key string, parent *html.Node, field StructField) {
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

	parent.AppendChild(n)
	h.decorator.NumberField(n)
}

func (h *htmlEncoder) buildNumberField(v reflect.Value, key string, parent *html.Node, field StructField) {
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
				Val: fmt.Sprintf("%v", v.Interface()),
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
	}

	parent.AppendChild(n)
	h.decorator.NumberField(n)
}

func (h *htmlEncoder) buildStringField(v reflect.Value, key string, parent *html.Node, field StructField) {
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

		parent.AppendChild(n)
		h.decorator.TextareaField(n)
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
					Val: field.Type(typField()),
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

		parent.AppendChild(n)
		h.decorator.TextField(n)
	}
}

func (h *htmlEncoder) buildBoolField(v reflect.Value, key string, parent *html.Node) {
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

	parent.AppendChild(n)
	h.decorator.CheckboxField(n)
}

func (h *htmlEncoder) buildSelectField(s Select, key string, parent *html.Node, field StructField) {
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

	for _, opt := range s.SelectOptions() {
		o := &html.Node{
			Type: html.ElementNode,
			Data: "option",
			Attr: []html.Attribute{
				{
					Key: "value",
					Val: fmt.Sprintf("%v", opt.Value),
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

		sel.AppendChild(o)
	}

	parent.AppendChild(sel)
	h.decorator.SelectField(sel)
}

func (h *htmlEncoder) buildRadioButtons(r Radio, key string, parent *html.Node, field StructField) {
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
					Val: fmt.Sprintf("%v", opt.Value),
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

		h.decorator.Label(label)
		h.decorator.RadioButton(radio)
	}

	parent.AppendChild(div)
}

func (h *htmlEncoder) formElementName(label string) string {
	return strings.Join(strings.Split(label, fieldSeparator)[2:], fieldSeparator)
}

func (h *htmlEncoder) buildLabel(label string, parent *html.Node, field StructField) {
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
	h.decorator.Label(n)
}

func (h *htmlEncoder) buildHelpText(label string, parent *html.Node, field StructField) {
	n := &html.Node{
		Type: html.ElementNode,
		Data: "div",
	}

	n.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: field.GetHelpText(),
	})

	parent.AppendChild(n)
	h.decorator.HelpText(n)
}
