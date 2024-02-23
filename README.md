# nuke

A memory arena implementation for Go.

[![Go Report Card](https://goreportcard.com/badge/github.com/ortuman/nuke?style=flat-square)](https://goreportcard.com/report/github.com/ortuman/nuke)
[![Godoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/ortuman/nuke)
[![Releases](https://img.shields.io/github/release/ortuman/nuke/all.svg?style=flat-square)](https://github.com/ortuman/nuke/releases)
[![LICENSE](https://img.shields.io/github/license/ortuman/nuke.svg?style=flat-square)](https://github.com/ortuman/nuke/blob/master/LICENSE)

<div align="center">
    <a href="#">
      <img src="./logo/logo-0.png">
    </a>
</div>

## Getting Started

### Installation

```sh
go get -u github.com/ortuman/nuke
```

### Usage Example

```go
package main

import (
	"github.com/ortuman/nuke"
)

type Foo struct { A int }

func main() {
	// Initialize a new memory arena with a slab size of 256KB 
	// and a max memory size of 20MB.
	arena := nuke.NewSlabArena(256*1024, 20*1024*1024)
	
	// Allocate a new object of type Foo.
	fooRef := nuke.New[Foo](arena)
	
	// Allocate a Foo slice with a capacity of 10 elements.
	fooSlice := nuke.MakeSlice[Foo](arena, 0, 10)
	
	// Append 20 elements to the slice allocating 
	// the required extra memory from the arena.
	for i := 0; i < 20; i++ {
            fooSlice = nuke.SliceAppend(arena, fooSlice, Foo{A: i})
	}
	
	// ...

	// When done, reset the arena (releasing slab buffer memory).
	arena.Reset(true)
	
	// From here on, any arena reference is invalid.
	// ...
}
```

Additionally, we can inject a memory arena as part of a context, with the purpose of being used throughout the lifecycle of certain operations, such as an HTTP request.

```go
func httpHandler(w http.ResponseWriter, r *http.Request) {
    // Inject memory arena into request context.
    arena := nuke.NewSlabArena(64*1024, 1024*1024)
    defer arena.Reset(true)
	
    ctx := nuke.InjectContextArena(r.Context(), arena)
    processRequest(ctx)
    
    // ...
}

func processRequest(ctx context.Context) {
    // Get the memory arena from the context.
    arena := nuke.ExtractContextArena(ctx)
	
    // ...
}

func main() {
    http.HandleFunc("/", httpHandler) // Set the handler for the "/" route
    fmt.Println("Server is listening on port 8080...")
    http.ListenAndServe(":8080", nil) // Listen on port 8080
}
```

## Contributing

We welcome contributions from the community! If you'd like to contribute, please fork the repository, make your changes, and submit a pull request.

## Contact
If you have any questions, feedback or suggestions, please feel free to contact me at ortuman@gmail.com. I'm always open to feedback and would love to hear from you!

## License

This project is licensed under the terms of the Apache-2.0 license.