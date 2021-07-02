package formulate

import (
	"encoding/json"
	"net/url"
	"reflect"
	"strconv"
	"time"
)

// HTTPDecoder takes a set of url values and decodes them.
type HTTPDecoder struct {
	showConditions

	form url.Values
}

// NewDecoder creates a new HTTPDecoder.
func NewDecoder(form url.Values) *HTTPDecoder {
	return &HTTPDecoder{
		form:           form,
		showConditions: make(map[string]ShowConditionFunc),
	}
}

// Decode a the given values into a provided interface{}. Note that the underlying
// value must be a pointer.
func (h *HTTPDecoder) Decode(data interface{}) error {
	val := reflect.ValueOf(data)

	if val.Kind() != reflect.Ptr {
		panic("formulate: decode target must be pointer")
	}

	elem := val.Elem()

	if elem.Kind() != reflect.Struct {
		panic("formulate: decode target underlying value must be struct")
	}

	if decoder, ok := data.(CustomDecoder); ok {
		data, err := decoder.DecodeFormValue(h.form, "", nil)

		if err != nil {
			return err
		}

		if data.Kind() == reflect.Ptr {
			data = data.Elem()
		}

		elem.Set(data)

		return nil
	}

	return h.decode(elem, elem.Type().String())
}

func (h *HTTPDecoder) getFormValues(key string) []string {
	key = formElementName(key)

	var vals []string

	if formValues, ok := h.form[key]; ok {
		vals = formValues
	}

	return vals
}

func (h *HTTPDecoder) decode(val reflect.Value, key string) error {
	if val.CanInterface() {
		switch a := val.Interface().(type) {
		case CustomDecoder:
			decodedFormVal, err := a.DecodeFormValue(h.form, key, h.getFormValues(key))

			if err != nil {
				return err
			}

			if !decodedFormVal.IsValid() {
				return nil
			}

			val.Set(decodedFormVal)

			return nil
		case time.Time:
			values := h.getFormValues(key)

			var t time.Time

			if len(values) > 0 {
				var err error

				t, err = time.Parse(timeFormat, values[0])

				if err != nil {
					return err
				}
			}

			val.Set(reflect.ValueOf(t))

			return nil
		}
	}

	values := h.getFormValues(key)

	var formValue string

	if len(values) > 0 {
		formValue = values[0]
	}

	switch val.Kind() {
	case reflect.Struct:
		// recurse over the fields
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			fieldType := val.Type().Field(i)
			structField := StructField{fieldType}

			if structField.Hidden(h.showConditions) {
				// hidden fields will not be in the form, so don't decode them.
				continue
			}

			err := h.decode(field, key+fieldSeparator+fieldType.Name)

			if err != nil {
				return err
			}
		}
		return nil
	case reflect.Ptr:
		// dereference ptr, decode again
		if val.IsNil() && val.CanAddr() {
			val.Set(reflect.New(val.Type().Elem()))
		}

		return h.decode(val.Elem(), key)
	case reflect.String:
		val.SetString(formValue)
		return nil
	case reflect.Float64, reflect.Float32:
		var f float64
		var err error

		if formValue != "" {
			f, err = strconv.ParseFloat(formValue, 64)

			if err != nil {
				return err
			}
		}

		val.SetFloat(f)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var i int64
		var err error

		if formValue != "" {
			i, err = strconv.ParseInt(formValue, 10, 0)

			if err != nil {
				return err
			}
		}

		val.SetInt(i)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var i uint64
		var err error

		if formValue != "" {
			i, err = strconv.ParseUint(formValue, 10, 0)

			if err != nil {
				return err
			}
		}

		val.SetUint(i)
		return nil
	case reflect.Bool:
		if formValue == "on" {
			val.SetBool(true)
		} else {
			var i int64
			var err error

			if formValue != "" {
				i, err = strconv.ParseInt(formValue, 10, 0)

				if err != nil {
					return err
				}
			}

			val.SetBool(i == 1)
		}

		return nil
	case reflect.Map, reflect.Slice, reflect.Array:
		i := reflect.New(val.Type())

		if formValue == "" {
			val.Set(i.Elem())
			return nil
		}

		if err := json.Unmarshal([]byte(formValue), i.Interface()); err != nil {
			return err
		}

		val.Set(i.Elem())
		return nil
	default:
		return nil
	}
}
