package formulate

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
)

func ExampleNewEncoder() {
	type Address struct {
		HouseName       string `help:"You can leave this blank."`
		AddressLine1    string
		AddressLine2    string
		Postcode        string
		TelephoneNumber Tel
		CountryCode     string `pattern:"[A-Za-z]{3}"`
	}

	buf := new(bytes.Buffer)

	address := Address{
		AddressLine1: "Fake Street",
	}

	encoder := NewEncoder(buf, nil, nil)
	encoder.SetFormat(true)

	if err := encoder.Encode(&address); err != nil {
		panic(err)
	}

	fmt.Println(buf.String())
	// Output:
	// <div>
	//   <fieldset>
	//     <div>
	//       <label for="HouseName">
	//         House Name
	//       </label>
	//       <div>
	//         <input type="text" name="HouseName" id="HouseName" value=""/>
	//         <div>
	//           You can leave this blank.
	//         </div>
	//       </div>
	//     </div>
	//     <div>
	//       <label for="AddressLine1">
	//         Address Line 1
	//       </label>
	//       <div>
	//         <input type="text" name="AddressLine1" id="AddressLine1" value="Fake Street"/>
	//         <div></div>
	//       </div>
	//     </div>
	//     <div>
	//       <label for="AddressLine2">
	//         Address Line 2
	//       </label>
	//       <div>
	//         <input type="text" name="AddressLine2" id="AddressLine2" value=""/>
	//         <div></div>
	//       </div>
	//     </div>
	//     <div>
	//       <label for="Postcode">
	//         Postcode
	//       </label>
	//       <div>
	//         <input type="text" name="Postcode" id="Postcode" value=""/>
	//         <div></div>
	//       </div>
	//     </div>
	//     <div>
	//       <label for="TelephoneNumber">
	//         Telephone Number
	//       </label>
	//       <div>
	//         <input type="tel" name="TelephoneNumber" id="TelephoneNumber" value=""/>
	//         <div></div>
	//       </div>
	//     </div>
	//     <div>
	//       <label for="CountryCode">
	//         Country Code
	//       </label>
	//       <div>
	//         <input type="text" name="CountryCode" id="CountryCode" value="" pattern="[A-Za-z]{3}"/>
	//         <div></div>
	//       </div>
	//     </div>
	//   </fieldset>
	// </div>
}

func ExampleNewDecoder() {
	type Address struct {
		HouseName       string `help:"You can leave this blank."`
		AddressLine1    string
		AddressLine2    string
		Postcode        string
		TelephoneNumber Tel
		CountryCode     string `pattern:"[A-Za-z]{3}" validators:"countryCode"`

		EmptyStruct struct {
			Foo string `show:"-"`
		}
	}

	// formValues - usually these would come from *http.Request.Form!
	formValues := url.Values{
		"HouseName":       {"1 Example Road"},
		"AddressLine1":    {"Fake Town"},
		"AddressLine2":    {"Fake City"},
		"Postcode":        {"Postcode"},
		"TelephoneNumber": {"012345678910"},
		"CountryCode":     {"GBR"},
	}

	var address Address

	decoder := NewDecoder(formValues)
	decoder.AddValidators(countryCodeValidator{})
	decoder.SetValueOnValidationError(true)

	if err := decoder.Decode(&address); err != nil {
		panic(err)
	}

	fmt.Printf("House Name: %s\n", address.HouseName)
	fmt.Printf("Line 1: %s\n", address.AddressLine1)
	fmt.Printf("Line 2: %s\n", address.AddressLine2)
	fmt.Printf("Postcode: %s\n", address.Postcode)
	fmt.Printf("Telephone: %s\n", address.TelephoneNumber)
	fmt.Printf("CountryCode: %s\n", address.CountryCode)
	// Output:
	// House Name: 1 Example Road
	// Line 1: Fake Town
	// Line 2: Fake City
	// Postcode: Postcode
	// Telephone: 012345678910
	// CountryCode: GBR
}

func ExampleFormulate() {
	type Address struct {
		HouseName       string `help:"You can leave this blank."`
		AddressLine1    string
		AddressLine2    string
		Postcode        string
		TelephoneNumber Tel
		CountryCode     string `pattern:"[A-Za-z]{3}" validators:"countryCode"`
	}

	buildEncoder := func(r *http.Request, w io.Writer) *HTMLEncoder {
		enc := NewEncoder(w, r, nil)
		enc.SetFormat(true)

		return enc
	}

	buildDecoder := func(r *http.Request, values url.Values) *HTTPDecoder {
		dec := NewDecoder(values)
		dec.SetValueOnValidationError(true)
		dec.AddValidators(countryCodeValidator{})

		return dec
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		var addressForm Address

		encodedForm, save, err := Formulate(r, &addressForm, buildEncoder, buildDecoder)

		if err == nil && save {
			// save the form here
			http.Redirect(w, r, "/", http.StatusFound)
		} else if err != nil {
			http.Error(w, "Bad Form", http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "text/html")
		_, _ = w.Write([]byte(encodedForm))
	}

	// for example purposes only.
	srv := httptest.NewServer(http.HandlerFunc(handler))
	defer srv.Close()

	resp, err := http.Get(srv.URL)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)

	fmt.Println(string(b))

	// Output:
	// <div>
	//   <fieldset>
	//     <div>
	//       <label for="HouseName">
	//         House Name
	//       </label>
	//       <div>
	//         <input type="text" name="HouseName" id="HouseName" value=""/>
	//         <div>
	//           You can leave this blank.
	//         </div>
	//       </div>
	//     </div>
	//     <div>
	//       <label for="AddressLine1">
	//         Address Line 1
	//       </label>
	//       <div>
	//         <input type="text" name="AddressLine1" id="AddressLine1" value=""/>
	//         <div></div>
	//       </div>
	//     </div>
	//     <div>
	//       <label for="AddressLine2">
	//         Address Line 2
	//       </label>
	//       <div>
	//         <input type="text" name="AddressLine2" id="AddressLine2" value=""/>
	//         <div></div>
	//       </div>
	//     </div>
	//     <div>
	//       <label for="Postcode">
	//         Postcode
	//       </label>
	//       <div>
	//         <input type="text" name="Postcode" id="Postcode" value=""/>
	//         <div></div>
	//       </div>
	//     </div>
	//     <div>
	//       <label for="TelephoneNumber">
	//         Telephone Number
	//       </label>
	//       <div>
	//         <input type="tel" name="TelephoneNumber" id="TelephoneNumber" value=""/>
	//         <div></div>
	//       </div>
	//     </div>
	//     <div>
	//       <label for="CountryCode">
	//         Country Code
	//       </label>
	//       <div>
	//         <input type="text" name="CountryCode" id="CountryCode" value="" pattern="[A-Za-z]{3}"/>
	//         <div></div>
	//       </div>
	//     </div>
	//   </fieldset>
	// </div>
}
