package opzione

// Simple is an optional type best for value types. For pointer types,
// it only checks whether the single pointer it stores is nil, with no guards
// towards nested pointers. For nested pointers, consider using Chained.
//
//	a := 10
//	p1 := &a
//	p2 := &p1
//
//	opt := SimpleSome(p2)
//	p1 = nil
//	println(opt.IsNone()) // false
type Simple[T interface{}] struct {
	empty  bool
	ptrtyp bool
	v      *T
}

// SimpleSome constructs a Simple optional with a value. It panics if v is
// a nil pointer.
func SimpleSome[T interface{}](v T) *Simple[T] {
	ptr, ok := isptr(v)
	if ok {
		if ptr == nil {
			panic("nil pointer cannot be used to construct Some")
		}
	}
	return &Simple[T]{ptrtyp: ok, v: &v}
}

// SimpleNone constructs a Simple optional with no value.
func SimpleNone[T interface{}]() *Simple[T] {
	var v T
	_, ok := isptr(v)
	return &Simple[T]{empty: true, ptrtyp: ok}
}

func (s *Simple[T]) IsNone() bool {
	return s.empty || s.v == nil
}

func (s *Simple[T]) Value() (t T, err error) {
	if s.IsNone() {
		return t, ErrNoneOptional
	}
	return *s.v, nil
}

func (s *Simple[T]) Must() T {
	if s.IsNone() {
		panic(ErrNoneOptional)
	}
	return *s.v
}

func (s *Simple[T]) Swap(v T) (t T) {
	if !s.IsNone() {
		t = *s.v
	}

	if s.ptrtyp {
		ptr, _ := isptr(v)
		if ptr == nil {
			panic("nil pointer cannot be used to construct Some")
		}
	}

	s.v, s.empty = &v, false
	return
}

func (s *Simple[T]) Take() (t T, err error) {
	if s.IsNone() {
		return t, ErrNoneOptional
	}
	t = *s.v
	s.v, s.empty = nil, true
	return t, nil
}

func (s *Simple[T]) With(f func(T)) {
	if !s.IsNone() {
		f(*s.v)
	}
}
