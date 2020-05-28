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
}

type htmlEncoder struct {
	n *html.Node
	w io.Writer

	decorator Decorator
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
		w:         w,
		n:         n,
		decorator: decorator,
	}
}

func (h *htmlEncoder) Encode(i interface{}) error {
	v := reflect.ValueOf(i)

	h.recurse(v, v.Type().String(), StructField{}, h.n)

	return html.Render(h.w, h.n)
}

func (h *htmlEncoder) recurse(v reflect.Value, key string, field StructField, parent *html.Node) {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() && v.CanAddr() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		h.recurse(v.Elem(), key, field, parent)
		return
	case reflect.Struct:
		switch v.Interface().(type) {
		case time.Time, Select, Radio:
			h.buildField(v, key, field, parent)
			return
		}

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
	if !v.IsValid() || field.Hidden() {
		return
	}

	rowElement := &html.Node{
		Type: html.ElementNode,
		Data: "div",
	}

	h.decorator.Row(rowElement)

	h.buildLabel(key, rowElement, field)

	defer parent.AppendChild(rowElement)

	switch a := v.Interface().(type) {
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
		h.buildNumberField(v, key, rowElement, field)
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
				Val: h.formElementName(key),
			},
			{
				Key: "id",
				Val: h.formElementName(key),
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

	h.decorator.NumberField(n)

	parent.AppendChild(n)
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
				Val: h.formElementName(key),
			},
			{
				Key: "id",
				Val: h.formElementName(key),
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

	h.decorator.NumberField(n)

	parent.AppendChild(n)
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
					Val: h.formElementName(key),
				},
				{
					Key: "id",
					Val: h.formElementName(key),
				},
			},
		}

		n.AppendChild(&html.Node{
			Type: html.TextNode,
			Data: v.String(),
		})

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
					Val: h.formElementName(key),
				},
				{
					Key: "id",
					Val: h.formElementName(key),
				},
				{
					Key: "value",
					Val: v.String(),
				},
			},
		}

		h.decorator.TextField(n)
	}

	parent.AppendChild(n)
}

func (h *htmlEncoder) buildBoolField(v reflect.Value, key string, parent *html.Node) {
	val := "0"

	if v.Bool() {
		val = "1"
	}

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
				Val: h.formElementName(key),
			},
			{
				Key: "id",
				Val: h.formElementName(key),
			},
			{
				Key: "value",
				Val: val,
			},
		},
	}

	h.decorator.CheckboxField(n)

	parent.AppendChild(n)
}

func (h *htmlEncoder) buildSelectField(s Select, key string, parent *html.Node, field StructField) {
	sel := &html.Node{
		Type: html.ElementNode,
		Data: "select",
		Attr: []html.Attribute{
			{
				Key: "name",
				Val: h.formElementName(key),
			},
			{
				Key: "id",
				Val: h.formElementName(key),
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

		if opt.Checked {
			o.Attr = append(o.Attr, html.Attribute{Key: "selected"})
		}

		o.AppendChild(&html.Node{
			Type: html.TextNode,
			Data: opt.Label,
		})

		sel.AppendChild(o)
	}

	h.decorator.SelectField(sel)

	parent.AppendChild(sel)
}

func (h *htmlEncoder) buildRadioButtons(r Radio, key string, parent *html.Node, field StructField) {
	elemName := h.formElementName(key)

	div := &html.Node{
		Type: html.ElementNode,
		Data: "div",
		Attr: []html.Attribute{
			{
				Key: "id",
				Val: elemName,
			},
		},
	}

	for i, opt := range r.RadioOptions() {
		id := fmt.Sprintf("%s%d", elemName, i)

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
					Val: elemName,
				},
			},
		}

		if opt.Disabled {
			radio.Attr = append(radio.Attr, html.Attribute{Key: "disabled"})
		}

		if opt.Checked {
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

		h.decorator.Label(label)
		h.decorator.RadioButton(radio)

		div.AppendChild(label)
		div.AppendChild(radio)
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
				Val: h.formElementName(label),
			},
		},
	}

	h.decorator.Label(n)

	n.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: field.GetName(),
	})

	parent.AppendChild(n)
}
