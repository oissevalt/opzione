// Package opzione provides operations with optional values.
package opzione

import (
	"errors"
	"reflect"
)

var ErrNoneOptional = errors.New("optional value is none")

type Optional[T interface{}] interface {
	// IsNone reports whether the current optional is None.
	IsNone() bool

	// Value tries to obtain the contained value. If the optional is None,
	// it returns ErrNoneOptional.
	Value() (t T, err error)

	// Must obtains the contained value, and panics if the optional is None.
	Must() T

	// Swap swaps value v with the optional's contained value. If the
	// optional was None, it will become Some. Swap returns the original
	// value, with no guarantee that it will be present.
	Swap(v T) T

	// Take attempts to move out the optional's contained value, leaving a
	// None behind. If the optional is None, it returns ErrNoneOptional.
	Take() (*T, error)

	// With accepts a closure which will be executed with the optional's
	// contained value, if it is Some.
	With(func(T))

	// Assign assigns the optional's contained value to *p, if the optional
	// is not None.
	Assign(p **T) bool
}

// NewOptional returns an optional type, whose specific implementation
// depends on the generic type T and initial value v.
//
//   - SimpleSome: T is a value type or a simple pointer type, and t is not nil.
//     However, SimpleSome is created for nil slices because they are still valid
//     (can be called safely with cap, len, or append).
//   - SimpleNone: T is a simple pointer type, and t is a nil pointer.
//   - ChainedSome: T is a nested pointer type, and t is not nil.
//   - ChainedNone: T is a nested pointer type, and t is nil or deferences to nil.
func NewOptional[T any](v T) Optional[T] {
	val := reflect.ValueOf(v)
	if !isptrkind(val.Kind()) {
		return &Simple[T]{v: &v}
	}

	switch val.Kind() {
	case reflect.UnsafePointer:
		// Only responsible for the topmost reference.
		return &Simple[T]{v: &v, ptrtyp: true}
	case reflect.Pointer, reflect.Interface:
		// If v is a simple pointer, there is no need to resort to
		// reflection.
		if !isptrkind(val.Elem().Kind()) {
			return &Simple[T]{v: &v, ptrtyp: true}
		}
		return &Chained[T]{v: &v, ptrtyp: true, track: true}
	default:
		return &Simple[T]{v: &v, ptrtyp: true}
	}
}
