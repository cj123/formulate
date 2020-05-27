package formulate

type (
	Password string
	Email    string
	URL      string
	Tel      string
)

type Select interface {
	// Multiple indicates whether multiple options can be selected at once.
	SelectMultiple() bool

	// SelectOptions are the available options
	SelectOptions() []Option
}

type Option struct {
	Value interface{}
	Label string

	Disabled bool
	Checked  bool
}

type Radio interface {
	RadioOptions() []Option
}
