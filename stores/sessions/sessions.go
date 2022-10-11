package sessions

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/cj123/sessions"

	"github.com/cj123/formulate"
)

func init() {
	gob.Register([]formulate.ValidationError{})
}

// Store implements formulate.ValidationStore using a combination of both HTTP session and filesystem.
// ValidationErrors are stored in the HTTP session, but the FormValue is stored in the os.TempDir() in a
// JSON encoded blob, with the filename formulate_val_* where * is replaced by a random string. The filesystem
// storage is used as the session storage is limited to 4096 bytes in most browsers.
type Store struct {
	r             *http.Request
	w             http.ResponseWriter
	sessionsStore sessions.Store
	sessionName   string
}

// NewStore creates a session store for saving validation. The sessionName provided must be unique to each form instance.
func NewStore(r *http.Request, w http.ResponseWriter, store sessions.Store, sessionName string) *Store {
	return &Store{
		r:             r,
		w:             w,
		sessionsStore: store,
		sessionName:   sessionName,
	}
}

func (s *Store) getSession() (*sessions.Session, error) {
	x, err := s.sessionsStore.Get(s.r, s.sessionName)

	if err != nil {
		return nil, err
	}

	x.Options.SameSite = http.SameSiteLaxMode

	return x, nil
}

var ErrValidationErrorTypeAssertionFailed = errors.New("sessions: validation error type assertion failed")

func (s *Store) GetValidationErrors(field string) ([]formulate.ValidationError, error) {
	sess, err := s.getSession()

	if err != nil {
		return nil, err
	}

	vals, ok := sess.Values[field]

	if !ok {
		return nil, nil
	}

	validationErrors, ok := vals.([]formulate.ValidationError)

	if !ok {
		return nil, ErrValidationErrorTypeAssertionFailed
	}

	return validationErrors, nil
}

func (s *Store) AddValidationError(field string, validationError formulate.ValidationError) error {
	vals, err := s.GetValidationErrors(field)

	if err != nil {
		return err
	}

	vals = append(vals, validationError)

	sess, err := s.getSession()

	if err != nil {
		return err
	}

	sess.Values[field] = vals

	return s.sessionsStore.Save(s.r, s.w, sess)
}

func (s *Store) ClearValidationErrors() error {
	sess, err := s.getSession()

	if err != nil {
		return err
	}

	sess.Values = nil
	sess.Options.MaxAge = -1 // delete

	return s.sessionsStore.Save(s.r, s.w, sess)
}

func (s *Store) SetFormValue(i interface{}) error {
	sess, err := s.getSession()

	if err != nil {
		return err
	}

	filename, err := s.persistFormValue(i)

	if err != nil {
		return err
	}

	sess.Values["form_value"] = filename

	return s.sessionsStore.Save(s.r, s.w, sess)
}

var ErrInvalidValue = errors.New("sessions: invalid value")

func (s *Store) GetFormValue(out interface{}) (err error) {
	sess, err := s.getSession()

	if err != nil {
		return err
	}

	defer func() {
		delete(sess.Values, "form_value")

		saveErr := s.sessionsStore.Save(s.r, s.w, sess)

		if err == nil {
			err = saveErr
		}
	}()

	val, ok := sess.Values["form_value"]

	if !ok {
		return ErrInvalidValue
	}

	name, ok := val.(string)

	if !ok {
		return ErrInvalidValue
	}

	return s.readFormValue(name, out)
}

func (s *Store) persistFormValue(val interface{}) (name string, err error) {
	file, err := ioutil.TempFile("", "formulate_val_*")

	if err != nil {
		return "", err
	}

	defer file.Close()

	if err := json.NewEncoder(file).Encode(val); err != nil {
		return "", err
	}

	return file.Name(), nil
}

func (s *Store) readFormValue(name string, out interface{}) (err error) {
	file, err := os.Open(name)

	if err != nil {
		return err
	}

	defer func() {
		err = file.Close()

		if err == nil {
			err = os.Remove(name)
		}
	}()

	return json.NewDecoder(file).Decode(out)
}
