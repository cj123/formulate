package formulate

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestHTTPDecoder_Decode(t *testing.T) {
	description := `This is a long description about the customer

It spans multiple lines`

	vals := url.Values{
		"Name":                                   {"John Smith"},
		"Age":                                    {"25"},
		"Email":                                  {"john.smith@example.com"},
		"ConfirmedEmail":                         {"on"},
		"Description":                            {description},
		"Password":                               {"hunter2"},
		"Time":                                   {"2020-05-28T15:28"},
		"Pet":                                    {"moose"},
		"ContactMethod":                          {"email"},
		joinFields("Address", "HouseName"):       {"1 Example Road"},
		joinFields("Address", "AddressLine1"):    {"Fake Town"},
		joinFields("Address", "AddressLine2"):    {"Fake City"},
		joinFields("Address", "Postcode"):        {"Postcode"},
		joinFields("Address", "TelephoneNumber"): {"012345678910"},
		joinFields("Address", "Country"):         {"UK"},
		joinFields("EmbeddedStruct", "Type"):     {"4838374"},
		"TestMap":                                {`{"Foo": "Banana", "baz": "chocolate"}`},
		"FavouriteNumber":                        {"1.222"},
		"FavouriteFoods":                         {"burger", "pizza", "beans"},
		"CountryCode":                            {"GBR"},
		"Checkbox":                               {"0"},
		"HiddenField":                            {"Content"},
	}

	details := YourDetails{EmbeddedStruct: EmbeddedStruct{SomeMultiselect: []string{"cake"}}}

	t.Run("Validation passes", func(t *testing.T) {
		dec := NewDecoder(vals)
		dec.AddValidators(&minAgeValidator{min: 20})
		dec.AddValidators(countryCodeValidator{})

		if err := dec.Decode(&details); err != nil {
			t.Error(err)
			return
		}

		assertEquals(t, details.Name, "John Smith")
		assertEquals(t, details.Age, 25)
		assertEquals(t, details.Email, Email("john.smith@example.com"))
		assertEquals(t, details.ConfirmedEmail, true)
		assertEquals(t, details.Description, description)
		assertEquals(t, details.Password, Password("hunter2"))
		assertEquals(t, details.Time.Format(time.RFC3339), "2020-05-28T15:28:00Z")
		assertEquals(t, details.Pet, Pet("moose"))
		assertEquals(t, details.ContactMethod, ContactMethod("email"))
		assertEquals(t, details.Address.HouseName, "1 Example Road")
		assertEquals(t, details.Address.AddressLine1, "Fake Town")
		assertEquals(t, details.Address.AddressLine2, "Fake City")
		assertEquals(t, details.Address.Postcode, "Postcode")
		assertEquals(t, details.Address.TelephoneNumber, Tel("012345678910"))
		assertEquals(t, details.Address.Country, "UK")
		assertEquals(t, details.Type, uint32(4838374))
		assertEquals(t, details.FavouriteNumber, 1.222)
		assertEquals(t, len(details.SomeMultiselect), 0)
		assertEquals(t, len(details.FavouriteFoods), 3)
		assertEquals(t, len(details.EmptySliceTest), 0)
		// hidden field should not be decoded, as it is hidden by a show condition
		assertEquals(t, details.HiddenField, "")
	})

	t.Run("Validation fails", func(t *testing.T) {
		vals.Set("CountryCode", "uk")

		dec := NewDecoder(vals)
		store := NewMemoryValidationStore()
		dec.SetValidationStore(store)
		dec.SetValueOnValidationError(true)
		dec.AddValidators(&minAgeValidator{min: 20})
		dec.AddValidators(countryCodeValidator{})

		if err := dec.Decode(&details); err != ErrFormFailedValidation {
			t.Fail()
		}

		validationErrors, err := store.GetValidationErrors("CountryCode")

		if err != nil || len(validationErrors) != 1 {
			t.Fail()
		}
	})

	t.Run("Decode on non-ptr type", func(t *testing.T) {
		dec := NewDecoder(nil)

		out := struct{}{}

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic() on non-ptr type.")
			}
		}()

		_ = dec.Decode(out)
	})

	t.Run("Decode on non-struct type", func(t *testing.T) {
		dec := NewDecoder(nil)

		var out int

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic() on non-struct type.")
			}
		}()

		_ = dec.Decode(&out)
	})

	t.Run("Custom decoder", func(t *testing.T) {
		dec := NewDecoder(url.Values{"CustomValue": {"1", "7", "33"}})

		var out customDecoderTest

		if err := dec.Decode(&out); err != nil {
			t.Error(err)
		}

		if out[0] != 1 && out[1] != 7 && out[2] != 33 {
			t.Fail()
		}
	})
}

type customDecoderTest []int

func (c customDecoderTest) DecodeFormValue(form url.Values, name string, values []string) (reflect.Value, error) {
	var out customDecoderTest

	for _, value := range form["CustomValue"] {
		i, err := strconv.ParseInt(value, 10, 0)

		if err != nil {
			return reflect.Value{}, err
		}

		out = append(out, int(i))
	}

	return reflect.ValueOf(out), nil
}

type emptySlice []string

func (e emptySlice) DecodeFormValue(form url.Values, name string, values []string) (reflect.Value, error) {
	if len(values) == 0 {
		return reflect.Value{}, nil
	}

	return reflect.ValueOf(emptySlice(values)), nil
}

func assertEquals(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		return
	}

	t.Logf("failed to assert that '%v' == '%v'", a, b)
	t.Fail()
}

func joinFields(fields ...string) string {
	return strings.Join(fields, fieldSeparator)
}

type minAgeValidator struct {
	min  int
	form url.Values
}

func (m *minAgeValidator) Validate(val interface{}) (ok bool, message string) {
	switch b := val.(type) {
	case int64:
		if int(b) >= m.min {
			return true, ""
		}
		return false, fmt.Sprintf("You must be over %d!", m.min)
	default:
		return false, "invalid type"
	}
}

func (m *minAgeValidator) TagName() string {
	return fmt.Sprintf("minAge(%d)", m.min)
}

func (m *minAgeValidator) SetForm(form url.Values) {
	m.form = form
}

type countryCodeValidator struct{}

func (c countryCodeValidator) Validate(value interface{}) (ok bool, message string) {
	switch a := value.(type) {
	case string:
		if len(a) == 3 && strings.ToUpper(a) == a {
			return true, ""
		}
		return false, "Country codes must be 3 letters and uppercase"
	default:
		return false, "invalid type"
	}
}

func (c countryCodeValidator) TagName() string {
	return "countryCode"
}

func TestHTTPDecoder_SetValidationStore(t *testing.T) {
	t.Run("Valid store", func(t *testing.T) {
		dec := NewDecoder(nil)
		v := NewMemoryValidationStore()

		dec.SetValidationStore(v)

		assertEquals(t, v, dec.validationStore)
	})

	t.Run("Nil store", func(t *testing.T) {
		dec := NewDecoder(nil)
		current := dec.validationStore

		dec.SetValidationStore(nil)

		assertEquals(t, dec.validationStore, current)
	})
}
