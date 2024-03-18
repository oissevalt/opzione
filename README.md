# Opzione

_Opzione_ (Italian "option") is a Go library for optionals, with tracking of nested pointers. You can add it to your project using `go get`:

```shell
go get github.com/oissevalt/opzione
```

The package provides two basic optional types (containers), `Simple` and `Chained`. It works by dynamically checking whether the containers contain meaningful values, which in most cases means not being `nil`. An exception are slices, as `nil` slices are safe to work with. Value types will never be none, except when the optional is constructed by calling `XXXNone`, or its value is moved out with `Take`.

> [!WARNING]
> This package's content, and behaviour have not yet been stabilized. Bugs and vulnerabilities may also be present.

## Simple

`Simple` is the basic implementation of a generic container of optional value. However, for pointer types, `Simple` only checks if the pointer it stores is `nil`. Thus, for nested pointers, `IsNone` only reflects the nil-ness of the shallowest reference.

```go
package main

import (
	"fmt"

	"github.com/oissevalt/opzione"
)

func main() {
	// Background information: a pointer to nil isn't nil.
	number := 10
	numptr := &number

	optional := opzione.SimpleSome(&numptr)
	fmt.Println(optional.IsNone()) // false
	
	numptr = nil // but &numptr is NOT nil
	fmt.Println(optional.IsNone()) // false
}
```

In the above example, `&numptr` is ultimately deferenced to `nil`, which can still cause a runtime panic, despite `optional.IsNone()` confidently reports `false`.

Therefore, `Simple` is preferred when you only need to work with value types or simple pointers.

## Chained

`Chained` is another optional container. It tracks nested pointers using runtime reflection. This way, it is able to report whether the nested pointers deference to `nil`, or a pointer is modified to be `nil` afterwards.

```go
package main

import (
	"fmt"

	"github.com/oissevalt/opzione"
)

func main() {
	number := 10
	numptr := &number

	optional := opzione.ChainedSome(&numptr)
	fmt.Println(optional.IsNone()) // false
	
	numptr = nil
	fmt.Println(optional.IsNone()) // true
}
```

Note that the use of reflection can introduce additional time cost and memory usage, but best effort has been made to minimize such impact. According to benchmark (`opzione_test.go`), a sequence of operation on nested pointers took less than 200ns on average.

For value types and single pointers, `Chained` is expected to behave the same as `Simple`. `Chained` does _not_ track unsafe pointers, either, because they can be arbitrarily manipulated and interpreted; there is no stable way to monitor them.

## Optional

`Optional` is the general interface for users to define their own optional type implementation. Refer to documentation in the source code for more information.
