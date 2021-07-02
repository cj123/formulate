package formulate

import (
	"reflect"
	"strings"

	"github.com/fatih/camelcase"
)

// StructField is a wrapper around the reflect.StructField type. The rendering behavior of form elements is controlled
// by Struct Tags. The following tags are currently available:
//
//  - name (e.g. name:"Phone Number") - this overwrites the name used in the label. This value can be left empty.
//  - help (e.g. help:"Enter your phone number, including area code") - this text is displayed alongside the input field as a prompt.
//  - show (e.g. show:"adminOnly") - controls visibility of elements. See HTMLEncoder.AddShowCondition for more details. if "contents" is used, the field is shown and the parent fieldset (if any) will be omitted.
//  - type (e.g. type:"tel") - sets the HTML input "type" attribute
//  - elem (elem:"textarea") - used to specify that a text input should use a <textarea> rather than an input field.
//  - min (e.g. min:"0") - minimum value for number inputs
//  - max (e.g. max:"10") - maximum value for number inputs
//  - step (e.g. step:"0.1") - step size for number inputs
//  - pattern (e.g. pattern:"[a-z]+" - regex pattern for text inputs
//
// These can all be used in combination with one another in a struct field. A full example of the above types is:
//
//    type Address struct {
//        HouseNumber          int `min:"0" help:"Please enter your house number" name:"House Number (if any)"
//        AddressLine1         string
//        DeliveryInstructions string `elem:"textarea"`
//        CountryCode          string `pattern:"[A-Za-z]{3}"`
//    }
type StructField struct {
	reflect.StructField
}

// GetName returns the name of the StructField, taking into account tag name overrides.
func (sf StructField) GetName() string {
	tagName := sf.Tag.Get("name")

	if tagName != "" {
		return tagName
	}

	return camelCase(sf.Name)
}

// GetHelpText returns the help text for the field.
func (sf StructField) GetHelpText() string {
	return sf.Tag.Get("help")
}

// Hidden determines if a StructField is hidden based on the showConditions.
func (sf StructField) Hidden(showConditions map[string]ShowConditionFunc) bool {
	show := sf.Tag.Get("show")

	if conditionFunc, ok := showConditions[show]; ok {
		return !conditionFunc()
	}

	return show == "-"
}

func camelCase(s string) string {
	return strings.Join(camelcase.Split(s), " ")
}

// InputType returns the HTML <input> element type attribute
func (sf StructField) InputType(original string) string {
	t := sf.Tag.Get("type")

	if t != "" {
		return t
	}

	return original
}

// Elem returns the element to be used. Currently the only supported value is <textarea>.
// <input> will be used if not specified.
func (sf StructField) Elem() string {
	return sf.Tag.Get("elem")
}

// HasMin determines if a StructField has a minimum value
func (sf StructField) HasMin() bool {
	return sf.Tag.Get("min") != ""
}

// Min is the minimum value of the StructField
func (sf StructField) Min() string {
	return sf.Tag.Get("min")
}

// HasMax determines if a StructField has a maximum value
func (sf StructField) HasMax() bool {
	return sf.Tag.Get("max") != ""
}

// Max is the maximum value of the StructField
func (sf StructField) Max() string {
	return sf.Tag.Get("max")
}

// HasStep determines if a StructField has a step value
func (sf StructField) HasStep() bool {
	return sf.Tag.Get("step") != ""
}

// Step value of the StructField
func (sf StructField) Step() string {
	return sf.Tag.Get("step")
}

func (sf StructField) Pattern() string {
	return sf.Tag.Get("pattern")
}

func (sf StructField) Placeholder() string {
	return sf.Tag.Get("placeholder")
}

func (sf StructField) Required() bool {
	return sf.Tag.Get("required") == "true"
}

// BuildFieldset determines whether a given struct should be inside its own fieldset. Use the Struct Tag
// show:"contents" to indicate that a fieldset should not be built for this struct.
func (sf StructField) BuildFieldset() bool {
	if sf.Tag.Get("show") == "contents" {
		return false
	}

	return !sf.StructField.Anonymous
}

// ShowConditionFunc is a function which determines whether or not to show a form element. See: HTMLEncoder.AddShowCondition
type ShowConditionFunc func() bool

type showConditions map[string]ShowConditionFunc

// AddShowCondition allows you to determine visibility of certain form elements.
// For example, given the following struct:
//   type Example struct {
//     Name string
//     SecretOption bool `show:"adminOnly"`
//   }
// If you wanted to make the SecretOption field only show to admins, you would call AddShowCondition as follows:
//   AddShowCondition("adminOnly", func() bool {
//      // some code that determines if we are 'admin'
//   })
// You can add multiple ShowConditions, but they must have different keys.
// Note: ShowConditions should be added to both the Encoder and Decoder.
func (s showConditions) AddShowCondition(key string, fn ShowConditionFunc) {
	s[key] = fn
}
