package formulate

import (
	"encoding/json"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// HTTPDecoder takes a set of url values and decodes them.
type HTTPDecoder struct {
	form url.Values
}

// NewDecoder creates a new HTTPDecoder.
func NewDecoder(form url.Values) *HTTPDecoder {
	return &HTTPDecoder{
		form: form,
	}
}

// Decode a the given values into a provided interface{}. Note that the underlying
// value must be a pointer.
func (h *HTTPDecoder) Decode(data interface{}) error {
	val := reflect.ValueOf(data)

	if val.Kind() != reflect.Ptr {
		panic("formulate: decode target must be pointer")
	}

	if decoder, ok := data.(CustomDecoder); ok {
		data, err := decoder.DecodeFormValue(h.form, "", nil)

		if err != nil {
			return err
		}

		val.Elem().Set(data)

		return nil
	}

	urlValues := make(url.Values)
	v := reflect.ValueOf(data).Elem()

	err := h.recurse(v, v.Type().String(), &urlValues)

	if err != nil {
		return err
	}

	for name, vals := range h.form {
		urlValues[name] = vals
	}

	for name, vals := range urlValues {
		if err := h.assignFieldValues(val.Elem(), name, vals); err != nil {
			return err
		}
	}

	return nil
}

func (h *HTTPDecoder) recurse(v reflect.Value, key string, urlValues *url.Values) error {
	switch v.Interface().(type) {
	case time.Time, Select, RadioList, CustomEncoder:
		name := formElementName(key)
		urlValues.Set(name, "")

		return nil
	}

	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() && v.CanAddr() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		return h.recurse(v.Elem(), key, urlValues)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			err := h.recurse(v.Field(i), key+"."+v.Type().Field(i).Name, urlValues)

			if err != nil {
				return err
			}
		}
		return nil
	default:
		name := formElementName(key)
		urlValues.Set(name, "")

		return nil
	}
}

const fieldSeparator = "."

func (h *HTTPDecoder) assignFieldValues(val reflect.Value, formName string, formValues []string) error {
	parts := strings.Split(formName, fieldSeparator)
	field := val.FieldByName(parts[0])

	if !field.CanSet() || len(formValues) == 0 {
		return nil
	}

	formValue := formValues[0]

	switch a := field.Interface().(type) {
	case CustomDecoder:
		val, err := a.DecodeFormValue(h.form, formName, formValues)

		if err != nil {
			return err
		}

		field.Set(val)

		return nil
	case time.Time:
		t, err := time.Parse(timeFormat, formValue)

		if err != nil {
			return err
		}

		field.Set(reflect.ValueOf(t))

		return nil
	}

	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			v := reflect.New(field.Type().Elem())

			field.Set(v)
		}

		field = field.Elem()
	}

	switch field.Kind() {
	case reflect.Struct:
		return h.assignFieldValues(field, strings.Join(parts[1:], fieldSeparator), formValues)
	case reflect.String:
		field.SetString(formValue)
		return nil
	case reflect.Float64, reflect.Float32:
		i, err := strconv.ParseFloat(formValue, 64)

		if err != nil {
			return err
		}

		field.SetFloat(i)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(formValue, 10, 0)

		if err != nil {
			return err
		}

		field.SetInt(i)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(formValue, 10, 0)

		if err != nil {
			return err
		}

		field.SetUint(i)
		return nil
	case reflect.Bool:
		if formValue == "on" {
			field.SetBool(true)
		} else {
			i, err := strconv.ParseInt(formValue, 10, 0)

			if err != nil {
				return err
			}

			field.SetBool(i == 1)
		}

		return nil
	case reflect.Map, reflect.Slice, reflect.Array:
		i := reflect.New(field.Type())

		if formValue == "" {
			field.Set(i.Elem())
			return nil
		}

		if err := json.Unmarshal([]byte(formValue), i.Interface()); err != nil {
			return err
		}

		field.Set(i.Elem())
		return nil
	default:
		// @TODO warning?
		// panic("formulate: unknown kind: " + field.Kind().String())
	}

	return nil
}
