package opzione

import (
	"os"
	"testing"
	"time"
)

// Interface assertions
var _ Optional[int] = &Option[int]{}

func BenchmarkNestedPointer(b *testing.B) {
	var model struct {
		name      string
		count     int
		addresses []string
		body      struct {
			content string
			ref     time.Time
		}
		kvs map[int]string
	}
	model.name = "model"
	model.count = 13
	model.addresses = []string{"_1", "_2", "_4"}
	model.body = struct {
		content string
		ref     time.Time
	}{os.TempDir(), time.Now()}
	model.kvs = map[int]string{2: "ad", 4: "bc"}

	for i := 0; i < b.N; i++ {
		ref := &model
		ref2 := &model

		optional := Some(&ref)
		if optional.IsNone() {
			b.Fatal("Unexpected None")
		}

		val := optional.Unwrap()
		if val == nil {
			b.Fatal("Unexpected nil")
		}

		ref = nil
		if !optional.IsNone() {
			b.Fatal("Unexpected Some")
		}

		optional.Swap(&ref2)
		if optional.IsNone() {
			b.Fatal("Unexpected None")
		}

		run := false
		optional.With(func(s **struct {
			name      string
			count     int
			addresses []string
			body      struct {
				content string
				ref     time.Time
			}
			kvs map[int]string
		}) {
			run = true
		})
		if !run {
			b.Fatal("Closure not run")
		}
	}
}

func TestValueTypes(t *testing.T) {
	option := Some(12)
	if option.IsNone() {
		t.Error("Unexpected None")
	}

	var (
		value int
		err   error
	)

	value = option.Unwrap()
	if value != 12 {
		t.Error("Unexpected value:", value)
	}

	swapped := option.Swap(24)
	value = option.Unwrap()
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
		option := Some[*os.File](nil)
		_ = option
	}, true)

	option := Some(file)
	if option.IsNone() {
		t.Error("Unexpected None")
	}

	file2 := option.Unwrap()
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
		optional := Some[**int](nil)
		_ = optional
	}, true)

	ShouldPanic(t, func() {
		ptr := &nilptr
		optional := Some(ptr)
		_ = optional
	}, true)

	optional := Some(&numptr)
	if optional.IsNone() {
		t.Error("Unexpected None")
	}

	var (
		value **int
		err   error
	)

	value = optional.Unwrap()
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

	optional := Some(noneSlice)
	if optional.IsNone() {
		t.Error("Unexpected None")
	}

	slice := optional.Unwrap()
	slice = append(slice, 1)
	optional.Swap(slice)

	optional.With(func(slice []int) {
		t.Log(slice)
	})

	noneSlice = nil

	optional2 := Some(&noneSlice)
	if optional.IsNone() {
		t.Error("Unexpected None")
	}

	slice2 := optional2.Unwrap()
	*slice2 = append(*slice2, 15)

	optional2.With(func(slice2 *[]int) {
		t.Log(*slice2)
	})
}

func TestPointerTypes(t *testing.T) {
	ShouldPanic(t, func() {
		var m map[int]int
		_ = Some(m)
	}, true)

	m := make(map[int]int)
	mapOptional := Some(m)

	m[1] = 2
	m2 := mapOptional.Unwrap()
	for k, v := range m2 {
		t.Log(k, v)
	}

	// ------------

	ShouldPanic(t, func() {
		var opt Optional[int] = nil
		_ = Some(opt)
	}, true)

	var opt Optional[int] = None[int]()
	interfaceOptional := Some(opt)

	opt2 := interfaceOptional.Unwrap()
	if !opt2.IsNone() {
		t.Error("Unexpected Some")
	}

	// ------------

	ShouldPanic(t, func() {
		var ch chan int = nil
		_ = Some(ch)
	}, true)

	ch := make(chan int)
	go func() {
		ch <- 15
	}()

	chOptional := Some(ch)
	if chOptional.IsNone() {
		t.Error("Unexpected None")
	}

	chOptional.With(func(c chan int) {
		t.Log("Received message from c:", <-c)
	})

	c := chOptional.Unwrap()
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
