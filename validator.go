package formulate

import (
	"errors"
	"net/url"
)

// Validator is an interface that allows individual form fields to be validated as part of the Decode phase of a formulate
// form. Validators may be reused for multiple form fields.
type Validator interface {
	// Validate a given value. If the value passes your validation, return true, and no message.
	// If the value fails validation, return false and a validation message.
	Validate(value interface{}) (ok bool, message string)
	// TagName is the name of the validator. This must match the tag name used in the struct tag on the form field.
	// For example for a field with a tag: `validators:"email"`, the TagName returned here must be "email".
	TagName() string
}

// FormAwareValidator is a Validator that is aware of the full form that was posted. This can be used for
// validation that requires knowledge of other form values.
type FormAwareValidator interface {
	Validator

	SetForm(form url.Values)
}

// ErrFormFailedValidation is returned if any form fields did not pass validation.
var ErrFormFailedValidation = errors.New("formulate: form failed validation")

// ValidatorKey is used to match the Validator's TagName against that on a StructField.
type ValidatorKey string

// ValidationError is an error generated by the validation process. Each field may have multiple validation errors.
type ValidationError struct {
	// Error is the error message returned by the Validator.Validate method.
	Error string

	// Value is the value which failed validation.
	Value interface{}
}

// ValidationStore is a data store for the validation errors
type ValidationStore interface {
	// GetValidationErrors returns the errors for a given field.
	GetValidationErrors(field string) []ValidationError
}

type nilValidationStore struct{}

func (n nilValidationStore) GetValidationErrors(_ string) []ValidationError {
	return nil
}