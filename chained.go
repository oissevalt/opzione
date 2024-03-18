package opzione

import (
	"reflect"
)

// Chained is an optional type which not only checks if the stored value
// is present, but also tracks nested references and their changes.
//
//	a := 10
//	p := &a
//	opt := ChainedSome(&p)
//
//	p = nil
//	println(opt.IsNone()) // true
//
// The tracking achieved using reflection, which may introduce overhead compared
// with Simple, but best effort has been made to keep the use of reflection
// to minimum.
//
// For value types, Chained skips reflection-powered nil checks, and is expected
// to behave the same as Simple. Chained does not track unsafe pointers, either,
// as they can be manipulated and interpreted arbitrarily.
type Chained[T any] struct {
	v      *T
	ptrtyp bool
	track  bool
}

// ChainedSome constructs a Chained optional with value. It panics if v is
// a nil pointer, or a nested pointer to nil, with nil slices being an exception.
func ChainedSome[T any](v T) *Chained[T] {
	val, ok := isptr(v)
	if ok {
		if isnil(val) {
			panic("nil pointer cannot be used to construct Some")
		}
		switch val.Kind() {
		case reflect.UnsafePointer:
			// Only responsible for the topmost reference.
			return &Chained[T]{v: &v, ptrtyp: true, track: false}
		case reflect.Pointer, reflect.Interface:
			// If v is a simple pointer, there is no need to resort to
			// reflection.
			tr := isptrkind(val.Elem().Kind())
			return &Chained[T]{v: &v, ptrtyp: true, track: tr}
		default:
			return &Chained[T]{v: &v, ptrtyp: true, track: true}
		}
	}
	return &Chained[T]{v: &v, ptrtyp: false, track: false}
}

// ChainedNone constructs an optional with no value.
func ChainedNone[T any]() *Chained[T] {
	var t T
	val, ok := isptr(t)
	if ok {
		switch val.Kind() {
		case reflect.UnsafePointer:
			return &Chained[T]{ptrtyp: true, track: false}
		case reflect.Pointer, reflect.Interface:
			tr := isptrkind(val.Elem().Kind())
			return &Chained[T]{ptrtyp: true, track: tr}
		default:
			return &Chained[T]{ptrtyp: true, track: true}
		}
	}
	return &Chained[T]{track: false}
}

// IsNone reports whether the current optional contains no value, merely
// a nil pointer, or nested pointers to a nil reference.
func (c *Chained[T]) IsNone() bool {
	if c.v == nil {
		return true
	}
	if c.ptrtyp {
		val := reflect.ValueOf(*c.v)
		if c.track {
			return isnil(val)
		}
		return val.IsNil()
	}
	return false
}

// Value attempts to retrieve the contained value. If the optional contains no value,
// is a nil pointer, or nested pointers to nil, it will return ErrNoneOptional.
func (c *Chained[T]) Value() (t T, err error) {
	if c.IsNone() {
		return t, ErrNoneOptional
	}
	return *c.v, nil
}

// Must returns the contained value, panicking if the optional is None.
func (c *Chained[T]) Must() T {
	if c.IsNone() {
		panic(ErrNoneOptional)
	}
	return *c.v
}

// Swap swaps the contained value with v, returning the original value. If v is
// a nil pointer, the current optional will be set to None. Whether the
// returned value is valid is not guaranteed; if the optional is previously None,
// it can be the zero value of the type, or nil.
func (c *Chained[T]) Swap(v T) (t T) {
	if !c.IsNone() {
		t = *c.v
	}
	// The value is accepted anyway, in case of pointer loss.
	//  var v importantNilType
	//  p := &v
	//  opt.Swap(p) // not good to lose reference to v
	c.v = &v
	return
}

// Take moves the inner value out, leaving the optional in a None state.
// It returns a reference to the contained value, if any. Should the optional
// previously be None, ErrNoneOptional is returned.
func (c *Chained[T]) Take() (*T, error) {
	if c.IsNone() {
		return nil, ErrNoneOptional
	}
	p := c.v
	c.v = nil
	return p, nil
}

// With executes the given closure with the contained value, if it is not None.
func (c *Chained[T]) With(f func(T)) {
	if !c.IsNone() {
		f(*c.v)
	}
}

// Assign assigns the inner value of the optional to *p, if the optional is
// not None. It returns a boolean indicating whether an assignment is made.
func (c *Chained[T]) Assign(p **T) bool {
	if c.IsNone() {
		return false
	}
	*p = c.v
	return true
}

func isptr[T any](t T) (reflect.Value, bool) {
	val := reflect.ValueOf(t)
	if !val.IsValid() {
		panic("cannot determine t; invalid value detected")
	}
	return val, isptrkind(val.Kind())
}

func isptrkind(kind reflect.Kind) bool {
	return kind == reflect.UnsafePointer ||
		kind == reflect.Pointer ||
		kind == reflect.Func ||
		kind == reflect.Map ||
		kind == reflect.Chan ||
		kind == reflect.Interface
}

func isnil(val reflect.Value) bool {
	if !val.IsValid() {
		// val is constructed from empty Value{}, nil, or is corrupted.
		return true
	}

	switch val.Kind() {
	case reflect.UnsafePointer:
		// An unsafe pointer can be anything; the package is only responsible
		// for checking the shallowest reference.
		return val.UnsafePointer() == nil
	case reflect.Pointer:
		elem := val.Elem()
		if !elem.IsValid() {
			// The pointer dereferences to nil; p := &i where i is nil.
			return true
		}
		// Continue this process with the pointed object.
		return isnil(elem)
	case reflect.Func, reflect.Map, reflect.Chan, reflect.Interface:
		// These are pointer-like types. They can be nil and calling a nil
		// value may trigger a runtime panic.
		return val.IsNil()
	case reflect.Slice:
		// A nil slice is safe to use. In the context of this package, we
		// don't consider it purely "nil" as opposed to a pointer.
		return false
	default:
		// Value types; cannot be nil.
		return false
	}
}
