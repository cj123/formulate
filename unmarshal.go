package formulate

import "net/http"

type Unmarshaler interface {
	Unmarshal(r *http.Request, val interface{}) error
}
