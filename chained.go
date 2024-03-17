package opzione

import (
	"reflect"
	"unsafe"
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
	inner    *T
	checkptr bool
}

// ChainedSome constructs a Chained optional with value. It panics if v is
// a nil pointer, or a nested pointer to nil.
func ChainedSome[T any](v T) Optional[T] {
	ptr, ok := isptr(v)
	if ok {
		if isnil(reflect.TypeOf(v), ptr) {
			panic("nil pointer cannot be used to construct Some")
		}
	}
	return &Chained[T]{inner: &v, checkptr: ok}
}

// ChainedNone constructs an optional with no value.
func ChainedNone[T any]() Optional[T] {
	var defaultValue T
	_, ok := isptr(defaultValue)
	return &Chained[T]{empty: true, checkptr: ok}
}

func (s *Chained[T]) IsNone() bool {
	if s.empty || s.inner == nil {
		return true
	} else {
		if !s.checkptr {
			return false
		}
		ptr, ok := isptr(s.inner)
		if !ok {
			s.checkptr = false
		}
		return ok && isnil(reflect.TypeOf(s.inner), ptr)
	}
}

func (s *Chained[T]) Value() (t T, err error) {
	if s.IsNone() {
		return t, ErrNoneOptional
	}
	return *s.inner, nil
}

func (s *Chained[T]) Must() T {
	if s.IsNone() {
		panic(ErrNoneOptional)
	}
	return *s.inner
}

func (s *Chained[T]) Swap(v T) (t T) {
	if !s.IsNone() {
		t = *s.inner
	}

	if s.checkptr {
		ptr, ok := isptr(v)
		if ok {
			if ptr == nil {
				s.inner = nil
				s.empty = true
			} else {
				s.inner = &v
				s.empty = isnil(reflect.TypeOf(v), ptr)
			}
		}
		s.checkptr = ok
	} else {
		s.inner = &v
		s.empty = false
	}

	return
}

func (s *Chained[T]) Take() (t T, err error) {
	if s.IsNone() {
		return t, ErrNoneOptional
	}
	p := s.inner
	s.inner, s.empty = nil, true
	return *p, nil
}

func (s *Chained[T]) With(f func(T)) {
	if !s.IsNone() {
		f(*s.inner)
	}
}

func isptr[T any](v T) (unsafe.Pointer, bool) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Pointer {
		return val.UnsafePointer(), true
	}
	return nil, false
}

func isnil(typ reflect.Type, p unsafe.Pointer) bool {
	if p == nil {
		return true
	}

	val := reflect.NewAt(typ, p).Elem()
	switch val.Kind() {
	case reflect.Pointer:
		pointed := val.Elem()
		if !pointed.IsValid() {
			return true
		}
		if pointed.Kind() != reflect.Pointer {
			return false
		}
		nestptr := pointed.Elem()
		if !nestptr.IsValid() {
			return true
		}
		if nestptr.Kind() != reflect.Pointer {
			return false
		}
		return isnil(nestptr.Type(), pointed.UnsafePointer())
	case reflect.Invalid:
		return true
	default:
		return false
	}
}
