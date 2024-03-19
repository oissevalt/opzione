# Opzione

_Opzione_ (Italian "option") is a Go library for optionals, with tracking of nested pointers. You can add it to your project using `go get`:

```shell
go get github.com/oissevalt/opzione
```

The package provides the optional type (container), `Option`. It works by dynamically checking whether the containers contain meaningful values, which in most cases means not being `nil`. An exception are slices, as `nil` slices are safe to work with. Value types will never be considered none, except when the optional is constructed by calling `None`, or its value is moved out with `Take`.

> [!WARNING]
> This package's content, and behaviour have not yet been stabilized. Bugs and vulnerabilities may also be present.

## Option

For pointer types, `Option` does not only checks if the pointer it stores is nil, but also tracks nested pointers using runtime reflection. This way, it is able to report whether some nested pointers deference to `nil`, or a pointer is modified to be `nil` after `Option` wraps it.

```go
package main

import (
	"fmt"

	"github.com/oissevalt/opzione"
)

func main() {
	number := 10
	numptr := &number

	option := opzione.Some(&numptr)
	fmt.Println(option.IsNone()) // false
	
	numptr = nil
	fmt.Println(option.IsNone()) // true
}
```

Note that the use of reflection can introduce additional time cost and memory usage, but best effort has been made to minimize such impact. According to benchmark (`opzione_test.go`), a sequence of operation on nested pointers took less than 200ns on average.

For value types and single pointers, `Option` does not enable tracking, and only checks the shallowest reference. It does _not_ track unsafe pointers, either, because they can be arbitrarily manipulated and interpreted; there is no stable way to monitor them.

## Optional

`Optional` is the general interface for users to define their own optional type implementation. Refer to documentation in the source code for more information.
