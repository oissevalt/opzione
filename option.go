package opzione

import (
	"reflect"
)

// Option is an optional type which not only checks if the stored value
// is present, but also tracks nested references and their changes.
//
//	a := 10
//	p := &a
//	opt := Some(&p)
//
//	p = nil
//	println(opt.IsNone()) // true
//
// The tracking achieved using reflection, which may introduce overhead,
// but best effort has been made to keep the use of reflection to minimum.
//
// For value types and single pointers, Option skips recursive nil checks.
// Option does not track unsafe pointers, either, as they can be manipulated
// and interpreted arbitrarily.
type Option[T any] struct {
	v       *T
	ptrtyp  bool
	track   bool
	validfn func(T) bool
}

// Validate adds custom validation logic when deciding whether the Option's
// inner value is meaningful or not. The value will be considered "none" if
// f returns true. The validation function is executed only after all nil
// checks are done.
func (o *Option[T]) Validate(f func(T) bool) {
	o.validfn = f
}

// IsNone reports whether the Option contains no value, or contains merely
// a nil pointer or nested pointers to a nil reference.
func (o *Option[T]) IsNone() bool {
	if o.v == nil {
		return true
	}

	ok := false
	if o.ptrtyp {
		val := reflect.ValueOf(*o.v)
		if o.track {
			ok = isnil(val)
		}
		ok = val.IsNil()
	}
	if ok {
		return true
	}

	if o.validfn != nil {
		return o.validfn(*o.v)
	}
	return false
}

// Value attempts to retrieve the contained value. If the Option contains no value,
// is a nil pointer or nested pointers to nil, it will return ErrNoneOptional.
func (o *Option[T]) Value() (t T, err error) {
	if o.IsNone() {
		return t, ErrNoneOptional
	}
	return *o.v, nil
}

// Unwrap returns the contained value, panicking if the Option contains no
// meaningful value.
func (o *Option[T]) Unwrap() T {
	if o.IsNone() {
		panic(ErrNoneOptional)
	}
	return *o.v
}

// Swap swaps the contained value with v, returning the original value. If v is
// a nil pointer or dereferences to nil, the Option will be put in a "none" state
// such that subsequent calls to IsNone will return true. Whether the returned
// value is valid is not guaranteed; if the optional previously contains no
// meaningful value, it can be the zero value of the type, or nil.
func (o *Option[T]) Swap(v T) (t T) {
	t = *o.v
	o.v = &v
	return
}

// Take moves the inner value out, leaving the optional in a "none" state such
// that subsequent calls to IsNone returns true. It returns a reference to the
// contained value, if any. Should the optional previously contains no meaningful
// value, ErrNoneOptional is returned; to forcibly move out an invalid pointer
// or value, consider calling Swap with nil.
func (o *Option[T]) Take() (*T, error) {
	if o.IsNone() {
		return nil, ErrNoneOptional
	}
	p := o.v
	o.v = nil
	return p, nil
}

// With executes the given closure, if the Option contains a meaningful value,
// with the contained value.
func (o *Option[T]) With(f func(T)) {
	if !o.IsNone() {
		f(*o.v)
	}
}

// WithNone executes the given closure only if the Option contains no value.
func (o *Option[T]) WithNone(f func()) {
	if o.IsNone() {
		f()
	}
}

// Assign assigns the inner value of the Option to *p, if it contains meaningful
// value. It returns a boolean indicating whether an assignment is made.
func (o *Option[T]) Assign(p **T) bool {
	if o.IsNone() {
		return false
	}
	*p = o.v
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
