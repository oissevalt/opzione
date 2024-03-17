package opzione

import "errors"

var ErrNoneOptional = errors.New("optional value is none")

type Optional[T interface{}] interface {
	// IsNone reports whether the current optional is None.
	IsNone() bool

	// Value tries to obtain the inner value. If the optional is None,
	// it returns ErrNoneOptional.
	Value() (t T, err error)

	// Must obtains the inner value, and panics if the optional is None.
	Must() T

	// Swap swaps value v with the optional's inner value. If the
	// optional was None, it will become Some. Swap returns the original
	// value, with no guarantee that it will be present.
	Swap(v T) T

	// Take attempts to move out the optional's inner value, leaving a
	// None behind. If the optional is None, it returns ErrNoneOptional.
	Take() (*T, error)

	// With accepts a closure which will be executed with the optional's
	// inner value, if it is Some.
	With(func(T))
}
