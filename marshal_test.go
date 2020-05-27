package formulate

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/yosssi/gohtml"
)

type YourDetails struct {
	Name           string `name:"Full Name"`
	Age            int    `step:"1" min:"0"`
	Email          Email
	ConfirmedEmail bool
	Description    string `elem:"textarea"`
	Password       Password
	Time           time.Time
	Pet            Pet
	ContactMethod  ContactMethod

	Address *Address
}

type Address struct {
	HouseName       string
	AddressLine1    string
	AddressLine2    string
	Postcode        string
	TelephoneNumber Tel
	Country         string
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
			Checked:  p == "dog",
		},
		{
			Value:    "cat",
			Label:    "Cat",
			Disabled: false,
			Checked:  p == "cat",
		},
		{
			Value:    "hamster",
			Label:    "Hamster",
			Disabled: true,
			Checked:  p == "hamster",
		},
		{
			Value:    "ferret",
			Label:    "Ferret",
			Disabled: false,
			Checked:  p == "ferret",
		},
	}
}

type ContactMethod string

func (c ContactMethod) RadioOptions() []Option {
	return []Option{
		{
			Value:   "phone",
			Label:   "Phone",
			Checked: c == "phone",
		},
		{
			Value:   "email",
			Label:   "Email",
			Checked: c == "email",
		},
		{
			Value:   "post",
			Label:   "Post",
			Checked: c == "post",
		},
		{
			Value:    "carrier_pigeon",
			Label:    "Carrier Pigeon",
			Disabled: true,
			Checked:  c == "carrier_pigeon",
		},
	}
}

func TestHtmlMarshaller_Marshal(t *testing.T) {
	buf := new(bytes.Buffer)
	m := NewHTMLMarshaller(buf, bootstrapDecorator{})

	if err := m.Marshal(YourDetails{
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
	}); err != nil {
		t.Error(err)
	}

	fmt.Println(gohtml.Format(buf.String()))
}
