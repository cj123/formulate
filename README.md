formulate [![Godoc][GodocSVG]][GodocURL]
=========

a HTML form builder and HTTP request to struct parser.

### Example (Formulate)

```go
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
	_, _ = w.Write(encodedForm)
}

// example validator
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
```

### Example (Encoder)

```go
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
```

Output:

```html
<div>
  <fieldset>
    <div>
      <label for="HouseName">
        House Name
      </label>
      <div>
        <input type="text" name="HouseName" id="HouseName" value=""/>
        <div>You can leave this blank.</div>
      </div>
    </div>
    <div>
      <label for="AddressLine1">
        Address Line 1
      </label>
      <div>
        <input type="text" name="AddressLine1" id="AddressLine1" value="Fake Street"/>
        <div></div>
      </div>
    </div>
    <div>
      <label for="AddressLine2">
        Address Line 2
      </label>
      <div>
        <input type="text" name="AddressLine2" id="AddressLine2" value=""/>
        <div></div>
      </div>
    </div>
    <div>
      <label for="Postcode">
        Postcode
      </label>
      <div>
        <input type="text" name="Postcode" id="Postcode" value=""/>
        <div></div>
      </div>
    </div>
    <div>
      <label for="TelephoneNumber">
        Telephone Number
      </label>
      <div>
        <input type="tel" name="TelephoneNumber" id="TelephoneNumber" value=""/>
        <div></div>
      </div>
    </div>
    <div>
      <label for="CountryCode">
        Country Code
      </label>
      <div>
        <input type="text" name="CountryCode" id="CountryCode" value="" pattern="[A-Za-z]{3}"/>
        <div></div>
      </div>
    </div>
  </fieldset>
</div>
```

### Example (Decoder)

```go
type Address struct {
    HouseName       string `help:"You can leave this blank."`
    AddressLine1    string
    AddressLine2    string
    Postcode        string
    TelephoneNumber Tel
    CountryCode     string `pattern:"[A-Za-z]{3}"`
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

if err := decoder.Decode(&address); err != nil {
    panic(err)
}

fmt.Printf("House Name: %s\n", address.HouseName)
fmt.Printf("Line 1: %s\n", address.AddressLine1)
fmt.Printf("Line 2: %s\n", address.AddressLine2)
fmt.Printf("Postcode: %s\n", address.Postcode)
fmt.Printf("Telephone: %s\n", address.TelephoneNumber)
fmt.Printf("CountryCode: %s\n", address.CountryCode)
```

Output:

```
House Name: 1 Example Road
Line 1: Fake Town
Line 2: Fake City
Postcode: Postcode
Telephone: 012345678910
CountryCode: GBR
```

   [GodocSVG]: https://godoc.org/github.com/cj123/formulate?status.svg
   [GodocURL]: https://godoc.org/github.com/cj123/formulate