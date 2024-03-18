package opzione

import (
	"os"
	"testing"
)

// Interface assertions
var _ Optional[int] = &Simple[int]{}
var _ Optional[int] = &Chained[int]{}

func TestValueTypes(t *testing.T) {
	option := SimpleSome(12)
	if option.IsNone() {
		t.Error("Unexpected None")
	}

	var (
		value int
		err   error
	)

	value = option.Must()
	if value != 12 {
		t.Error("Unexpected value:", value)
	}

	swapped := option.Swap(24)
	value = option.Must()
	if swapped != 12 {
		t.Error("Unexpected swapped value:", swapped)
	}
	if value != 24 {
		t.Error("Unexpected value:", value)
	}

	_, _ = option.Take()
	_, err = option.Value()
	if err == nil {
		t.Error("Unexpected nil error")
	}
	if !option.IsNone() {
		t.Error("Unexpected Some")
	}

	option.With(func(_ int) {
		t.Fatal("Closure executed when it shouldn't")
	})

	run := false
	option.Swap(36)
	option.With(func(_ int) {
		run = true
	})
	if !run {
		t.Fatal("Closure not executed when it should")
	}

	var numptr *int
	option.Assign(&numptr)

	if *numptr != 36 {
		t.Error("Assign failed:", *numptr)
	}
}

func TestSimplePointers(t *testing.T) {
	file, err := os.Open("go.mod")
	if err != nil {
		t.Fatal("Error when preparing *os.File:", err)
	}
	defer file.Close()

	ShouldPanic(t, func() {
		option := SimpleSome[*os.File](nil)
		_ = option
	}, true)

	option := SimpleSome(file)
	if option.IsNone() {
		t.Error("Unexpected None")
	}

	file2 := option.Must()
	if file2.Name() != "go.mod" {
		t.Fatal("Unexpected file object:", file2.Name())
	}

	run := false
	option.With(func(_ *os.File) {
		run = true
	})
	if !run {
		t.Error("Closure not run when it should")
	}

	option.Swap(nil)
	if !option.IsNone() {
		t.Error("Unexpected Some")
	}
	option.Swap(file)

	file3, _ := option.Take()
	if *file3 != file {
		t.Error("Different references found:", *file3, file)
	}
}

func TestChainedOptional(t *testing.T) {
	number := 10
	numptr := &number
	nilptr := (**int)(nil)

	ShouldPanic(t, func() {
		optional := ChainedSome[**int](nil)
		_ = optional
	}, true)

	ShouldPanic(t, func() {
		ptr := &nilptr
		optional := ChainedSome(ptr)
		_ = optional
	}, true)

	optional := ChainedSome(&numptr)
	if optional.IsNone() {
		t.Error("Unexpected None")
	}

	var (
		value **int
		err   error
	)

	value = optional.Must()
	if **value != 10 {
		t.Error("Unexpected pointer:", **value)
	}

	*value = nil
	if !optional.IsNone() {
		t.Error("Unexpected some")
	}

	*value = &number
	if optional.IsNone() {
		t.Error("Unexpected None")
	}

	optional.Swap(nilptr)
	if !optional.IsNone() {
		t.Error("Unexpected Some")
	}

	_, err = optional.Value()
	if err == nil {
		t.Error("Unexpected nil error")
	}

	optional.Swap(&numptr)

	ptr := &nilptr
	optional.Assign(&ptr)
	if n := ***ptr; n != 10 {
		t.Error("Unexpected dereference")
	}
}

func TestSlices(t *testing.T) {
	noneSlice := []int(nil)

	optional := SimpleSome(noneSlice)
	if optional.IsNone() {
		t.Error("Unexpected None")
	}

	slice := optional.Must()
	slice = append(slice, 1)
	optional.Swap(slice)

	optional.With(func(slice []int) {
		t.Log(slice)
	})

	noneSlice = nil

	optional2 := ChainedSome(&noneSlice)
	if optional.IsNone() {
		t.Error("Unexpected None")
	}

	slice2 := optional2.Must()
	*slice2 = append(*slice2, 15)

	optional2.With(func(slice2 *[]int) {
		t.Log(*slice2)
	})
}

func TestPointerTypes(t *testing.T) {
	ShouldPanic(t, func() {
		var m map[int]int
		_ = SimpleSome(m)
	}, true)

	m := make(map[int]int)
	mapOptional := SimpleSome(m)

	m[1] = 2
	m2 := mapOptional.Must()
	for k, v := range m2 {
		t.Log(k, v)
	}

	// ------------

	ShouldPanic(t, func() {
		var opt Optional[int] = nil
		_ = ChainedSome(opt)
	}, true)

	var opt Optional[int] = SimpleNone[int]()
	interfaceOptional := ChainedSome(opt)

	opt2 := interfaceOptional.Must()
	if !opt2.IsNone() {
		t.Error("Unexpected Some")
	}

	// ------------

	ShouldPanic(t, func() {
		var ch chan int = nil
		_ = ChainedSome(ch)
	}, true)

	ch := make(chan int)
	go func() {
		ch <- 15
	}()

	chOptional := SimpleSome(ch)
	if chOptional.IsNone() {
		t.Error("Unexpected None")
	}

	chOptional.With(func(c chan int) {
		t.Log("Received message from c:", <-c)
	})

	c := chOptional.Must()
	go func() {
		t.Log("Received message from c:", <-c)
	}()
	c <- 28
}

func ShouldPanic(t *testing.T, fn func(), p bool) {
	defer func() {
		panicked := false
		if err := recover(); err != nil {
			panicked = true
			if !p {
				t.Fatal("Panicked when not expected:", err)
			}
			t.Log("Recovered from an expected panic")
		}
		if !panicked && p {
			t.Fatal("Panic did not happen when expected")
		}
	}()
	fn()
}
