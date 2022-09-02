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
//  - show (e.g. show:"adminOnly") - controls visibility of elements. See HTMLEncoder.AddShowCondition for more details.
//    If "contents" is used, the field is shown and the parent fieldset (if any) will be omitted.
//    If "fieldset" is used, anonymous structs will be built as fieldsets too, if their name is also set.
//  - type (e.g. type:"tel", type:"hidden") - sets the HTML input "type" attribute. type:"hidden" will be rendered without labels and help text.
//  - elem (elem:"textarea") - used to specify that a text input should use a <textarea> rather than an input field.
//  - min (e.g. min:"0") - minimum value for number inputs
//  - max (e.g. max:"10") - maximum value for number inputs
//  - step (e.g. step:"0.1") - step size for number inputs
//  - pattern (e.g. pattern:"[a-z]+" - regex pattern for text inputs
//  - required (true/false) - adds the required attribute to the element.
//  - placeholder (e.g. placeholder:"phone number") - indicates a placeholder for the element.
//  - validators (e.g. "email,notempty") - which registered Validators to use.
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

	// ValidationErrors are the errors present for the StructField. They are only set on an encode.
	ValidationErrors []ValidationError
}

// GetName returns the name of the StructField, taking into account tag name overrides.
func (sf StructField) GetName() string {
	tagName := sf.Tag.Get("name")

	if tagName == "-" {
		return ""
	}

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
	showTag := sf.Tag.Get("show")

	for _, tag := range strings.Split(showTag, ",") {
		if conditionFunc, ok := showConditions[tag]; ok {
			return !conditionFunc()
		}
	}

	return showTag == "-"
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

// Elem returns the element to be used. Currently, the only supported value is <textarea>.
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

// Pattern is the regex for the input field.
func (sf StructField) Pattern() string {
	return sf.Tag.Get("pattern")
}

// Placeholder defines the placeholder attribute for the input field
func (sf StructField) Placeholder() string {
	return sf.Tag.Get("placeholder")
}

// Required indicates that an input field must be filled in.
func (sf StructField) Required() bool {
	return sf.Tag.Get("required") == "true"
}

// BuildFieldset determines whether a given struct should be inside its own fieldset. Use the Struct Tag
// show:"contents" to indicate that a fieldset should not be built for this struct. Use show:"fieldset"
// to indicate that anonymous structs should be built in a fieldset.
func (sf StructField) BuildFieldset() bool {
	showTag := sf.Tag.Get("show")

	for _, tag := range strings.Split(showTag, ",") {
		if tag == "contents" {
			return false
		} else if tag == "fieldset" {
			// allow anonymous structs to be built in a fieldset
			return true
		}
	}

	return !sf.StructField.Anonymous
}

// Validators are the TagNames of the registered Validators. Multiple Validators may be specified, separated by a comma.
func (sf StructField) Validators() []ValidatorKey {
	split := strings.Split(sf.Tag.Get("validators"), ",")

	var keys []ValidatorKey

	for _, key := range split {
		keys = append(keys, ValidatorKey(key))
	}

	return keys
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
