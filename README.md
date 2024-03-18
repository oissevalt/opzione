# Opzione

_Opzione_ (Italian "option") is a Go library for optionals, with tracking of nested pointers. You can add it to your project using `go get`:

```shell
go get github.com/oissevalt/opzione
```

The package provides two basic optional types (containers), `Simple` and `Chained`.

> [!WARNING]  
> This package's content, and behaviour have not yet been stabilized. Bugs and vulnerabilities may also be present.

## Simple

`Simple` is the basic implementation of optional value. However, for pointer types, `Simple` only checks if the pointer it currently stores is `nil`. Thus, for nested pointers, calling `IsNone` only reflects the nil-ness of the shallowest reference.

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

In the above example, `&numptr` is ultimately deferenced to `nil`, which can still cause a runtime panic, especially when `IsNone` confidently reports `false`.

Therefore, `Simple` is preferred when you only need to work with value types or simple pointers.

## Chained

`Chained` tracks nested pointers and their changes using runtime reflection. This way, it is able to report whether the nested pointers deference to `nil`.

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

Note that the use of reflection can introduce additional operation time and memory usage, but best effort has been made to minimize such impact. According to benchmark (`opzione_test.go`), a sequence of operation on nested pointers took less than 200ns on average.

For value types and single pointers, `Chained` is expected to behave the same as `Simple`. `Chained` does _not_ track unsafe pointers, either, because they can be arbitrarily manipulated and interpreted; there is no stable way to monitor them.

## Optional

`Optional` is the general interface for users to define their own optional type implementation. Refer to documentation in the source code for more information.
