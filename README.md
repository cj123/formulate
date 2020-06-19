formulate [![Godoc][GodocSVG]][GodocURL]
=========

a HTML form builder and HTTP request to struct parser.

### Example (Encoder)

```go
type Address struct {
    HouseName       string `help:"You can leave this blank."`
    AddressLine1    string
    AddressLine2    string
    Postcode        string
    TelephoneNumber Tel
    Country         string
}

buf := new(bytes.Buffer)

address := Address{
    AddressLine1: "Fake Street",
}

encoder := NewEncoder(buf, nil)
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
      <label for="Country">
        Country
      </label>
      <div>
        <input type="text" name="Country" id="Country" value=""/>
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
    Country         string
}

// formValues - usually these would come from *http.Request.Form!
formValues := url.Values{
    "HouseName":       {"1 Example Road"},
    "AddressLine1":    {"Fake Town"},
    "AddressLine2":    {"Fake City"},
    "Postcode":        {"Postcode"},
    "TelephoneNumber": {"012345678910"},
    "Country":         {"UK"},
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
fmt.Printf("Country: %s\n", address.Country)
```

Output:

```
House Name: 1 Example Road
Line 1: Fake Town
Line 2: Fake City
Postcode: Postcode
Telephone: 012345678910
Country: UK
```

   [GodocSVG]: https://godoc.org/github.com/cj123/formulate?status.svg
   [GodocURL]: https://godoc.org/github.com/cj123/formulate