package opzione

import (
	"fmt"
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
// This is achieved using reflection, which may introduce overhead compared
// with Simple, but best effort has been made to keep the use of reflection
// to minimum. For value types, Chained behaves the same as Simple.
type Chained[T any] struct {
	empty    bool
	v        *T
	checkptr bool
}

// ChainedSome constructs a Chained optional with value. It panics if v is
// a nil pointer, or a nested pointer to nil, with nil slices being an exception.
func ChainedSome[T any](v T) *Chained[T] {
	val, ok := isptr(v)
	if !val.IsValid() {
		panic("nil pointer or other invalid value cannot be used to construct Some")
	}
	if ok {
		if val.IsNil() || dereftonil(val) {
			panic("nil pointer cannot be used to construct Some")
		}
	}
	return &Chained[T]{v: &v, checkptr: ok}
}

// ChainedNone constructs an optional with no value.
func ChainedNone[T any]() *Chained[T] {
	var defaultValue T
	_, ok := isptr(defaultValue)
	return &Chained[T]{empty: true, checkptr: ok}
}

// IsNone reports whether the current optional contains no value, merely
// a nil pointer, or nested pointers to a nil reference.
func (c *Chained[T]) IsNone() bool {
	if c.empty || c.v == nil {
		return true
	} else {
		if !c.checkptr {
			return false
		}
		val, ok := isptr(*c.v)
		if !ok {
			c.checkptr = false
		}
		return ok && (val.IsNil() || dereftonil(val))
	}
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

	if c.checkptr {
		val, ok := isptr(v)
		if ok {
			if val.IsNil() {
				c.v = nil
				c.empty = true
			} else {
				c.v = &v
				c.empty = dereftonil(val)
			}
		}
		c.checkptr = ok
	} else {
		c.v = &v
		c.empty = false
	}

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
	c.v, c.empty = nil, true
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

func isptr[T any](v T) (val reflect.Value, b bool) {
	val = reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Chan, reflect.Func, reflect.Pointer, reflect.Map, reflect.Interface:
		return val, true
	default:
		return
	}
}

func dereftonil(val reflect.Value) bool {
	switch val.Kind() {
	case reflect.Pointer:
		pointed := val.Elem()
		if !pointed.IsValid() {
			return true
		}
		switch pointed.Kind() {
		case reflect.Func, reflect.Map, reflect.Chan:
			return pointed.IsNil()
		case reflect.Interface:
			if !pointed.IsValid() {
				return true
			}
			fallthrough
		case reflect.Pointer:
			return dereftonil(pointed.Elem())
		default:
			return false
		}
	case reflect.Func, reflect.Map, reflect.Chan:
		fmt.Println("Type:", val.Kind(), "IsNil:", val.IsNil())
		return val.IsNil()
	case reflect.Interface:
		return !val.IsValid()
	case reflect.Invalid:
		return true
	default:
		return false
	}
}
