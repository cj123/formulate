package formulate

import (
	"net/url"
	"testing"
	"time"
)

func TestHttpDecoder_Decode(t *testing.T) {
	description := `This is a long description about the customer

It spans multiple lines`

	vals := url.Values{
		"Name":                    {"John Smith"},
		"Age":                     {"25"},
		"Email":                   {"john.smith@example.com"},
		"ConfirmedEmail":          {"on"},
		"Description":             {description},
		"Password":                {"hunter2"},
		"Time":                    {"2020-05-28T15:28"},
		"Pet":                     {"moose"},
		"ContactMethod":           {"email"},
		"Address.HouseName":       {"1 Example Road"},
		"Address.AddressLine1":    {"Fake Town"},
		"Address.AddressLine2":    {"Fake City"},
		"Address.Postcode":        {"Postcode"},
		"Address.TelephoneNumber": {"012345678910"},
		"Address.Country":         {"UK"},
		"EmbeddedStruct.Type":     {"4838374"},
		"TestMap":                 {`{"Foo": "Banana", "baz": "chocolate"}`},
	}

	var details YourDetails

	if err := NewDecoder(vals).Decode(&details); err != nil {
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
}

func assertEquals(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		return
	}

	t.Logf("failed to assert that '%v' == '%v'", a, b)
	t.Fail()
}
