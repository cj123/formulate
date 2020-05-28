package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cj123/formulate"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	mux := http.DefaultServeMux
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		d := YourDetails{}

		if r.Method == http.MethodPost {
			if err := r.ParseForm(); err != nil {
				panic(err)
			}

			decoder := formulate.NewDecoder()

			if err := decoder.Decode(r.Form, &d); err != nil {
				panic(err)
			}

			spew.Fdump(w, d)
			return
		}

		buf := new(bytes.Buffer)

		encoder := formulate.NewEncoder(buf, formulate.BootstrapDecorator{})

		if err := encoder.Encode(&d); err != nil {
			panic(err)
		}

		w.Header().Set("Content-Type", "text/html")

		fmt.Fprintf(w, `<html lang="en">
<head>
<title>Form!</title>
<link rel="stylesheet" type="text/css" href="https://stackpath.bootstrapcdn.com/bootstrap/4.5.0/css/bootstrap.min.css">
</head>
<body>
	<div class="container">
		<form action="/" method="POST">
			%s
	
			<button class="btn btn-success float-right" type="submit">Submit</button>
		</form>
	</div>
</body>
</html>`, buf.String())
	})

	log.Fatal(http.ListenAndServe("0.0.0.0:4556", mux))
}

type YourDetails struct {
	Name           string `name:"Full Name"`
	Age            int    `step:"1" min:"0"`
	Email          formulate.Email
	ConfirmedEmail bool
	Description    string `elem:"textarea"`
	Password       formulate.Password
	Time           time.Time
	Pet            Pet
	ContactMethod  ContactMethod

	Address *Address
}

type Address struct {
	EmbeddedStruct

	HouseName       string
	AddressLine1    string
	AddressLine2    string
	Postcode        string
	TelephoneNumber formulate.Tel
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

func (p Pet) SelectOptions() []formulate.Option {
	return []formulate.Option{
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

func (c ContactMethod) RadioOptions() []formulate.Option {
	return []formulate.Option{
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
