package formulate

import (
	"bytes"
	"testing"
	"time"
)

type YourDetails struct {
	EmbeddedStruct

	Name            string `name:"Full Name"`
	Age             int    `step:"1" min:"0"`
	Email           Email
	ConfirmedEmail  bool
	Description     string `elem:"textarea"`
	Password        Password
	Time            time.Time
	Pet             Pet
	ContactMethod   ContactMethod
	FavouriteNumber float64

	Address *Address

	TestMap map[string]string
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
	Variable string
	Type     uint32
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
	m := NewEncoder(buf, nil)
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
	}); err != nil {
		t.Error(err)
	}

	if err := m.Encode(&YourDetails{}); err != nil {
		panic(err)
	}

	//fmt.Println(buf.String())
}
