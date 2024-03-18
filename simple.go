package opzione

import "reflect"

// Simple is an optional type best for value types and simple pointers.
// For pointer types, it only checks whether the single pointer it stores
// is nil, with no guards towards nested pointers. For nested pointers,
// consider using Chained.
//
//	a := 10
//	p1 := &a
//	p2 := &p1
//
//	opt := SimpleSome(p2)
//	p1 = nil
//	println(opt.IsNone()) // false
type Simple[T interface{}] struct {
	ptrtyp bool
	v      *T
}

// SimpleSome constructs a Simple optional with a value. It panics if v is
// a nil pointer, with nil slices being an exception.
func SimpleSome[T interface{}](v T) *Simple[T] {
	val, ok := isptr(v)
	if ok {
		if val.Kind() == reflect.UnsafePointer {
			if val.UnsafePointer() == nil {
				panic("nil pointer cannot be used to construct Some")
			}
		} else if val.IsNil() {
			panic("nil pointer cannot be used to construct Some")
		}
	}
	return &Simple[T]{ptrtyp: ok, v: &v}
}

// SimpleNone constructs a Simple optional with no value.
func SimpleNone[T interface{}]() *Simple[T] {
	var v T
	_, ok := isptr(v)
	return &Simple[T]{ptrtyp: ok}
}

// IsNone reports whether the current optional contains no value, or merely
// a nil pointer.
func (s *Simple[T]) IsNone() bool {
	if s.v == nil {
		return true
	}
	if s.ptrtyp {
		return reflect.ValueOf(*s.v).IsNil()
	}
	return false
}

// Value attempts to retrieve the contained value. If the optional contains no value,
// it will return ErrNoneOptional.
func (s *Simple[T]) Value() (t T, err error) {
	if s.IsNone() {
		return t, ErrNoneOptional
	}
	return *s.v, nil
}

// Must returns the contained value, panicking if the optional is None.
func (s *Simple[T]) Must() T {
	if s.IsNone() {
		panic(ErrNoneOptional)
	}
	return *s.v
}

// Swap swaps the contained value with v, returning the original value. If v is
// a nil pointer, the current optional will be set to None. Whether the
// returned value is valid is not guaranteed; if the optional is previously None,
// it can be the zero value of the type, or nil.
func (s *Simple[T]) Swap(v T) (t T) {
	if !s.IsNone() {
		t = *s.v
	}
	s.v = &v
	return
}

// Take moves the inner value out, leaving the optional in a None state.
// It returns a reference to the contained value, if any. Should the optional
// previously be None, ErrNoneOptional is returned.
func (s *Simple[T]) Take() (*T, error) {
	if s.IsNone() {
		return nil, ErrNoneOptional
	}
	t := s.v
	s.v = nil
	return t, nil
}

// With executes the given closure with the contained value, if it is not None.
func (s *Simple[T]) With(f func(T)) {
	if !s.IsNone() {
		f(*s.v)
	}
}

// Assign assigns the inner value of the optional to *p, if the optional is
// not None. It returns a boolean indicating whether an assignment is made.
func (s *Simple[T]) Assign(p **T) bool {
	if s.IsNone() {
		return false
	}
	*p = s.v
	return true
}
