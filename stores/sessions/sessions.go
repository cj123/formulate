package sessions

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/cj123/formulate"
	"github.com/cj123/sessions"
)

func init() {
	gob.Register([]formulate.ValidationError{})
}

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

	b, err := json.Marshal(i)

	if err != nil {
		return err
	}

	sess.Values["value"] = b

	return s.sessionsStore.Save(s.r, s.w, sess)
}

var ErrInvalidValue = errors.New("sessions: invalid value")

func (s *Store) GetFormValue(out interface{}) (err error) {
	sess, err := s.getSession()

	if err != nil {
		return err
	}

	defer func() {
		delete(sess.Values, "value")

		saveErr := s.sessionsStore.Save(s.r, s.w, sess)

		if err == nil {
			err = saveErr
		}
	}()

	val, ok := sess.Values["value"]

	if !ok {
		return ErrInvalidValue
	}

	b, ok := val.([]byte)

	if !ok {
		return ErrInvalidValue
	}

	return json.Unmarshal(b, out)
}
