package formulate

import (
	"net/url"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestHttpDecoder_Decode(t *testing.T) {
	vals := url.Values{
		"Name":           {"John Smith"},
		"Age":            {"25"},
		"Email":          {"john.smith@example.com"},
		"ConfirmedEmail": {"on"},
		"Description": {`This is a long description about the customer

It spans multiple lines`},
		"Password":                    {"hunter2"},
		"Time":                        {"2020-05-28T15:28"},
		"Pet":                         {"moose"},
		"ContactMethod":               {"email"},
		"Address.HouseName":           {"1 Example Road"},
		"Address.AddressLine1":        {"Fake Town"},
		"Address.AddressLine2":        {"Fake City"},
		"Address.Postcode":            {"Postcode"},
		"Address.TelephoneNumber":     {"012345678910"},
		"Address.Country":             {"UK"},
		"Address.EmbeddedStruct.Type": {"4838374"},
		"TestMap":                     {`{"Foo": "Banana", "baz": "chocolate"}`},
	}

	var details YourDetails

	if err := NewDecoder(vals).Decode(&details); err != nil {
		t.Error(err)
	}

	spew.Dump(details)
}
