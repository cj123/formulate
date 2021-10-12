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

	validators                map[ValidatorKey]Validator
	validationStore           ValidationStore
	setValueOnValidationError bool
	numValidationErrors       int
}

// NewDecoder creates a new HTTPDecoder.
func NewDecoder(form url.Values) *HTTPDecoder {
	return &HTTPDecoder{
		showConditions: make(map[string]ShowConditionFunc),
		form:           form,

		validators:                make(map[ValidatorKey]Validator),
		validationStore:           NewMemoryValidationStore(),
		setValueOnValidationError: false,
	}
}

func (h *HTTPDecoder) SetValidationStore(v ValidationStore) {
	if v == nil {
		return
	}

	h.validationStore = v
}

// SetValueOnValidationError indicates whether a form value should be set in the form if there was a validation error on that value.
func (h *HTTPDecoder) SetValueOnValidationError(b bool) {
	h.setValueOnValidationError = b
}

// AddValidators registers Validators to the decoder.
func (h *HTTPDecoder) AddValidators(validators ...Validator) {
	for _, validator := range validators {
		h.validators[ValidatorKey(validator.TagName())] = validator
	}
}

func (h *HTTPDecoder) getValidators(keys []ValidatorKey) []Validator {
	var validators []Validator

	for _, key := range keys {
		validator, ok := h.validators[key]

		if !ok {
			continue
		}

		validators = append(validators, validator)
	}

	return validators
}

// Decode the given values into a provided interface{}. Note that the underlying
// value must be a pointer.
func (h *HTTPDecoder) Decode(data interface{}) error {
	val := reflect.ValueOf(data)

	if val.Kind() != reflect.Ptr {
		panic("formulate: decode target must be pointer")
	}

	elem := val.Elem()

	// look for FormAwareValidators and set the form before we decode the data.
	for _, validator := range h.validators {
		if formAwareValidator, ok := validator.(FormAwareValidator); ok {
			formAwareValidator.SetForm(h.form)
		}
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

	if elem.Kind() != reflect.Struct {
		panic("formulate: decode target underlying value must be struct")
	}

	if err := h.decode(elem, elem.Type().String(), nil); err != nil {
		return err
	}

	if h.numValidationErrors > 0 {
		if err := h.validationStore.SetFormValue(data); err != nil {
			return err
		}

		return ErrFormFailedValidation
	}

	return nil
}

func (h *HTTPDecoder) getFormValues(key string) []string {
	key = formElementName(key)

	var vals []string

	if formValues, ok := h.form[key]; ok {
		vals = formValues
	}

	return vals
}

func (h *HTTPDecoder) decode(val reflect.Value, key string, validators []Validator) error {
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

			if decodedFormVal.CanInterface() {
				i := decodedFormVal.Interface()

				if ok, err := h.passedValidation(key, i, validators); ok && err == nil {
					val.Set(decodedFormVal)
				} else if err != nil {
					return err
				}
			} else {
				// validation is not possible
				val.Set(decodedFormVal)
			}

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

			if ok, err := h.passedValidation(key, t, validators); ok && err == nil {
				val.Set(reflect.ValueOf(t))
			} else if err != nil {
				return err
			}

			return nil
		}
	}

	switch val.Kind() {
	case reflect.Struct:
		// recurse over the fields
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			fieldType := val.Type().Field(i)
			structField := StructField{StructField: fieldType}

			if structField.Hidden(h.showConditions) {
				// hidden fields will not be in the form, so don't decode them.
				continue
			}

			err := h.decode(field, key+fieldSeparator+fieldType.Name, h.getValidators(structField.Validators()))

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

		return h.decode(val.Elem(), key, validators)
	case reflect.Interface:
		n := reflect.New(val.Elem().Type())
		n.Elem().Set(val.Elem())

		if err := h.decode(n, key, validators); err != nil {
			return err
		}

		val.Set(n.Elem())

		return nil
	}

	values := h.getFormValues(key)

	if len(values) == 0 {
		// below we are dealing with concrete types that do not call decode recursively.
		// if there are no values in the form for these types, do not decode them. this
		// prevents 'default' values from being overwritten with empty values.
		return nil
	}

	formValue := values[0]

	switch val.Kind() {
	case reflect.String:
		if ok, err := h.passedValidation(key, formValue, validators); ok && err == nil {
			val.SetString(formValue)
		} else if err != nil {
			return err
		}

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

		if ok, err := h.passedValidation(key, f, validators); ok && err == nil {
			val.SetFloat(f)
		} else if err != nil {
			return err
		}

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

		if ok, err := h.passedValidation(key, i, validators); ok && err == nil {
			val.SetInt(i)
		} else if err != nil {
			return err
		}

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

		if ok, err := h.passedValidation(key, i, validators); ok && err == nil {
			val.SetUint(i)
		} else if err != nil {
			return err
		}

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

			b := i == 1

			if ok, err := h.passedValidation(key, b, validators); ok && err == nil {
				val.SetBool(b)
			} else if err != nil {
				return err
			}
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

func (h *HTTPDecoder) passedValidation(key string, value interface{}, validators []Validator) (bool, error) {
	ok := true

	for _, validator := range validators {
		valid, message := validator.Validate(value)

		if !valid {
			h.numValidationErrors++

			err := h.validationStore.AddValidationError(formElementName(key), ValidationError{
				Value: value,
				Error: message,
			})

			if err != nil {
				return ok, err
			}

			ok = false
		}
	}

	return ok || h.setValueOnValidationError, nil
}
