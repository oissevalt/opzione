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

	// Unwrap obtains the contained value, and panics if the optional is None.
	Unwrap() T

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

	// WithNone executes the given closure, if the optional currently contains
	// no value.
	WithNone(func())

	// Assign assigns the optional's contained value to *p, if the optional
	// is not None.
	Assign(p **T) bool
}

// Some constructs an Option with value. It panics if v is a nil pointer,
// or a nested pointer to nil, with nil slices being an exception.
func Some[T any](v T) *Option[T] {
	val, ok := isptr(v)
	if ok {
		if isnil(val) {
			panic("nil pointer cannot be used to construct Some")
		}
		switch val.Kind() {
		case reflect.UnsafePointer:
			// Only responsible for the topmost reference.
			return &Option[T]{v: &v, ptrtyp: true, track: false}
		case reflect.Pointer, reflect.Interface:
			// If v is a simple pointer, there is no need to resort to
			// reflection.
			tr := isptrkind(val.Elem().Kind())
			return &Option[T]{v: &v, ptrtyp: true, track: tr}
		default:
			return &Option[T]{v: &v, ptrtyp: true, track: true}
		}
	}
	return &Option[T]{v: &v, ptrtyp: false, track: false}
}

// None constructs an Option with no value.
func None[T any]() *Option[T] {
	var t T
	val, ok := isptr(t)
	if ok {
		switch val.Kind() {
		case reflect.UnsafePointer:
			return &Option[T]{ptrtyp: true, track: false}
		case reflect.Pointer, reflect.Interface:
			tr := isptrkind(val.Elem().Kind())
			return &Option[T]{ptrtyp: true, track: tr}
		default:
			return &Option[T]{ptrtyp: true, track: true}
		}
	}
	return &Option[T]{track: false}
}
