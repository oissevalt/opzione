package opzione

import (
	"fmt"
	"testing"
)

func BenchmarkOptionalReflection(b *testing.B) {
	number := 10
	pointer := &number

	for i := 0; i < b.N; i++ {
		optional := ChainedSome(&pointer)

		_ = optional.IsNone()
		_, _ = optional.Value()

		pointer = nil
		_ = optional.IsNone()
		_, _ = optional.Value()

		p2 := &number
		_ = optional.Swap(&p2)

		optional.With(func(i **int) {
			_ = **i + 1
		})

		pointer = &number
	}
}

func TestSimple_Example(t *testing.T) {
	a := 10
	p := &a

	opt := SimpleSome(&p)
	expect("TestSimple 1", t, opt.IsNone(), false)

	p = nil
	expect("TestSimple 2", t, opt.IsNone(), false)
}

func TestChained_Example(t *testing.T) {
	a := 10
	p := &a

	opt := ChainedSome(&p)
	expect("TestChained 1", t, opt.IsNone(), false)

	p = nil
	expect("TestChained 2", t, opt.IsNone(), true)
}

func TestValueOptional(t *testing.T) {
	someValue := SimpleSome[int](12)
	noneValue := SimpleNone[int]()

	expect("1", t, someValue.IsNone(), false)
	expect("2", t, noneValue.IsNone(), true)

	value, err := someValue.Value()
	expect("3", t, err, nil)
	expect("4", t, value, 12)

	value, err = noneValue.Value()
	expect("5", t, err, ErrNoneOptional)
	expect("6", t, value, 0)

	someValue.Swap(23)
	value, err = someValue.Value()
	expect("7", t, err, nil)
	expect("8", t, value, 23)

	swapped := noneValue.Swap(15)
	expect("9", t, noneValue.IsNone(), false)
	expect("9.1", t, swapped, 0)

	value, err = noneValue.Value()
	expect("10", t, err, nil)
	expect("11", t, value, 15)

	value2, err := someValue.Take()
	expect("12", t, err, nil)
	expect("13", t, *value2, 23)
	expect("14", t, someValue.IsNone(), true)
}

func TestPointerOptional(t *testing.T) {
	value1 := 12
	value2 := 24

	somePointer := ChainedSome[*int](&value1)
	nonePointer := ChainedNone[*int]()

	expect("1", t, somePointer.IsNone(), false)
	expect("2", t, nonePointer.IsNone(), true)

	value, err := somePointer.Value()
	expect("3", t, err, nil)
	expect("4", t, value, &value1)

	value, err = nonePointer.Value()
	expect("5", t, err, ErrNoneOptional)
	expect("6", t, value, nil)

	swapped := somePointer.Swap(&value2)
	expect("7.1", t, swapped, &value1)

	value, err = somePointer.Value()
	expect("7", t, err, nil)
	expect("8", t, value, &value2)

	nonePointer.Swap(&value1)
	expect("9", t, nonePointer.IsNone(), false)

	value, err = nonePointer.Value()
	expect("10", t, err, nil)
	expect("11", t, value, &value1)

	value3, err := somePointer.Take()
	expect("12", t, err, nil)
	expect("13", t, *value3, &value2)
	expect("14", t, somePointer.IsNone(), true)

	swapped = somePointer.Swap(nil)
	expect("7.1", t, swapped, nil)

	value, err = somePointer.Value()
	expect("7", t, err, ErrNoneOptional)
	expect("8", t, value, nil)

	intptr := &value2
	nested := &intptr

	t.Logf("intptr: %p, nested: %p, &nested: %p", intptr, nested, &nested)

	ptrptrOption := ChainedSome[***int](&nested)

	expect("15", t, ptrptrOption.IsNone(), false)

	valueptr, err := ptrptrOption.Value()
	expect("16", t, valueptr, &nested)
	expect("17", t, err, nil)

	fmt.Println("---------------")
	intptr = nil
	valueptr, err = ptrptrOption.Value()
	expect("18", t, ptrptrOption.IsNone(), true)
}

func expect[T comparable](tag string, t *testing.T, lhs T, rhs T) {
	if lhs != rhs {
		t.Errorf("%s\tlhs: %v, rhs: %v", tag, lhs, rhs)
		return
	}
	// t.Logf("%s passed", tag)
}
