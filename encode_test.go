package formulate

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/csrf"
)

type YourDetails struct {
	EmbeddedStruct

	Name            string `name:"Full Name"`
	Age             int    `step:"1" min:"0" validators:"minAge(20)"`
	Email           Email
	ConfirmedEmail  bool
	Description     string `elem:"textarea"`
	Password        Password
	Time            time.Time
	Pet             Pet
	ContactMethod   ContactMethod
	FavouriteNumber float64
	CountryCode     string `pattern:"[A-Za-z]{3}" validators:"countryCode"`
	FavouriteFoods  FoodSelect
	Checkbox        bool
	IgnoredField    string `show:"-"`
	HiddenInput     string `type:"hidden"`
	EmptyStruct     struct {
		Foo string `show:"-"`
	}

	Address *Address

	TestMap map[string]string

	_               string
	unexportedField string
}

type FoodSelect []string

func (f FoodSelect) SelectMultiple() bool {
	return true
}

func (f FoodSelect) SelectOptions() []Option {
	return []Option{
		{
			Value: "burger",
			Label: "burger",
		},
		{
			Value: "pizza",
			Label: "pizza",
		},
		{
			Value: "beans",
			Label: "beans",
		},
		{
			Value: "banana",
			Label: "banana",
		},
	}
}

func (f FoodSelect) DecodeFormValue(form url.Values, name string, values []string) (reflect.Value, error) {
	if len(values) == 0 {
		return reflect.Value{}, nil
	}

	return reflect.ValueOf(FoodSelect(values)), nil
}

type Address struct {
	HouseName       string `help:"You can leave this blank."`
	AddressLine1    string
	AddressLine2    string
	Postcode        string
	TelephoneNumber Tel
	Country         string
}

type EmbeddedStruct struct {
	Variable        string
	Type            uint32
	SomeMultiselect []string

	EmptySliceTest emptySlice
}

type Pet string

func (p Pet) SelectMultiple() bool {
	return false
}

func (p Pet) SelectOptions() []Option {
	return []Option{
		{
			Value:    "dog",
			Label:    "Dog",
			Disabled: false,
		},
		{
			Value:    "cat",
			Label:    "Cat",
			Disabled: false,
		},
		{
			Value:    "hamster",
			Label:    "Hamster",
			Disabled: true,
		},
		{
			Value:    "ferret",
			Label:    "Ferret",
			Disabled: false,
		},
	}
}

type ContactMethod string

func (c ContactMethod) RadioOptions() []Option {
	return []Option{
		{
			Value: "phone",
			Label: "Phone",
		},
		{
			Value: "email",
			Label: "Email",
		},
		{
			Value: "post",
			Label: "Post",
		},
		{
			Value:    "carrier_pigeon",
			Label:    "Carrier Pigeon",
			Disabled: true,
		},
	}
}

func TestHtmlEncoder_Encode(t *testing.T) {
	buf := new(bytes.Buffer)
	m := NewEncoder(buf, nil, nil)
	m.SetFormat(true)

	if err := m.Encode(&YourDetails{
		Name:           "Jane Doe",
		Age:            40,
		ConfirmedEmail: true,
		Password:       "hunter2",
		Description:    "This is a description of the customer",
		Time:           time.Now(),
		Pet:            "cat",
		ContactMethod:  "email",
		Address: &Address{
			HouseName:       "Fake House",
			AddressLine1:    "Fake Street",
			AddressLine2:    "Fake Town",
			Postcode:        "F4K3 T0WN",
			TelephoneNumber: "012345678910",
			Country:         "UK",
		},
		TestMap: map[string]string{
			"Foo": "foo",
			"Bar": "bar",
		},
		HiddenInput: "hidden-val",
	}); err != nil {
		t.Error(err)
	}

	if err := m.Encode(&YourDetails{}); err != nil {
		t.Error(err)
	}

	t.Run("Multi-select encoding set selected value automatically", func(t *testing.T) {
		type test struct {
			Food FoodSelect

			Number numberIndexedSelect
		}

		s := &test{Food: FoodSelect{"burger", "pizza"}, Number: numberIndexedSelect{1, 2}}

		buf := new(bytes.Buffer)
		m := NewEncoder(buf, nil, nil)
		m.SetFormat(true)

		if err := m.Encode(s); err != nil {
			t.Error(err)
		}

		expected := `<div>
  <fieldset>
    <div>
      <label for="Food">
        Food
      </label>
      <div>
        <select name="Food" id="Food" multiple="">
          <option value="burger" selected="">
            burger
          </option>
          <option value="pizza" selected="">
            pizza
          </option>
          <option value="beans">
            beans
          </option>
          <option value="banana">
            banana
          </option>
        </select>
        <div></div>
      </div>
    </div>
    <div>
      <label for="Number">
        Number
      </label>
      <div>
        <select name="Number" id="Number" multiple="">
          <option value="0">
            Zero
          </option>
          <option value="1" selected="">
            One
          </option>
          <option value="2" selected="">
            Two
          </option>
        </select>
        <div></div>
      </div>
    </div>
  </fieldset>
</div>`

		if expected != buf.String() {
			t.Fail()
		}
	})

	t.Run("Encoder adds CSRF protection input to form if enabled", func(t *testing.T) {
		mux := http.NewServeMux()
		p := csrf.Protect([]byte(""))

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			s := struct{}{}

			m := NewEncoder(w, r, nil)
			m.SetFormat(true)
			m.SetCSRFProtection(true)

			if err := m.Encode(s); err != nil {
				t.Error(err)
			}
		})

		srv := httptest.NewServer(p(mux))

		defer srv.Close()

		resp, err := http.Get(srv.URL)

		if err != nil {
			t.Error(err)
		}

		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			t.Error(err)
		}

		if !strings.Contains(string(b), `<input type="hidden" name="gorilla.csrf.Token"`) {
			t.Error("Expected gorilla CSRF token input in HTTP response")
		}
	})

	t.Run("Encoder with Show Conditions", func(t *testing.T) {
		type test struct {
			Name                    string
			AddressLine1            string `show:"visible"`
			AddressLine2            string `show:"visible,invisible"`
			PostCode                string `show:"invisible"`
			HiddenByGlobalCondition int
		}

		s := &test{}

		buf := new(bytes.Buffer)
		m := NewEncoder(buf, nil, nil)
		m.SetFormat(true)
		m.AddShowCondition("visible", func(field StructField) bool {
			return true
		})
		m.AddShowCondition("invisible", func(field StructField) bool {
			return false
		})
		m.AddGlobalShowCondition(func(field StructField) bool {
			return field.Name != "HiddenByGlobalCondition"
		})

		if err := m.Encode(s); err != nil {
			t.Error(err)
		}

		b := buf.String()

		if !strings.Contains(b, "AddressLine1") {
			t.Fail()
		}

		if strings.Contains(b, "PostCode") {
			t.Fail()
		}

		if strings.Contains(b, "AddressLine2") {
			t.Fail()
		}

		if strings.Contains(b, "HiddenByGlobalCondition") {
			t.Fail()
		}
	})
}

type numberIndexedSelect []int

func (n numberIndexedSelect) SelectMultiple() bool {
	return true
}

func (n numberIndexedSelect) SelectOptions() []Option {
	return []Option{
		{
			Value: 0,
			Label: "Zero",
		},
		{
			Value: 1,
			Label: "One",
		},
		{
			Value: 2,
			Label: "Two",
		},
	}
}
