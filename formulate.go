package formulate

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
	"net/url"
)

// Formulate is an all-in-one method for handling form encoding and decoding, including validation errors.
// This expects the form to be POST-ed to the same endpoint as the form is displayed on. If you require a custom
// implementation of the form handling (including on separate endpoints), this can be done with the
// HTMLEncoder.Encode and HTTPDecoder.Decode methods.
// The Formulate method overrides any ValidationStore already set and uses a MemoryValidationStore instead.
func Formulate(r *http.Request, data interface{}, encoderBuilder HTMLEncoderBuilder, decoderBuilder HTTPDecoderBuilder) (encodedForm template.HTML, passedValidation bool, err error) {
	var decoder *HTTPDecoder
	validationStore := NewMemoryValidationStore()

	if r.Method == http.MethodPost {
		if err = r.ParseForm(); err != nil {
			return "", passedValidation, err
		}

		decoder = decoderBuilder(r, r.Form)
		decoder.SetValidationStore(validationStore)

		err := decoder.Decode(data)

		if err == nil {
			passedValidation = true
		} else if err != ErrFormFailedValidation {
			return "", passedValidation, err
		}
	}

	buf := new(bytes.Buffer)

	encoder := encoderBuilder(r, buf)
	encoder.SetValidationStore(validationStore)

	if err := encoder.Encode(data); err != nil {
		return "", passedValidation, err
	}

	return template.HTML(buf.Bytes()), passedValidation, nil
}

// HTMLEncoderBuilder is a function that builds a HTMLEncoder given an io.Writer as the output.
// When used with the Formulate method, this allows for custom building of the encoder (including ShowConditions etc).
type HTMLEncoderBuilder func(r *http.Request, w io.Writer) *HTMLEncoder

// HTTPDecoderBuilder is a function that builds a HTTPDecoder given the form as the input.
// When used with the Formulate method, this allows for custom building of the decoder (including ShowConditions, Validators etc).
type HTTPDecoderBuilder func(r *http.Request, values url.Values) *HTTPDecoder
