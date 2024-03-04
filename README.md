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

## But wait, what is a memory arena?

A memory arena is a method of memory management where a large block of memory is allocated at once and portions of it are used to satisfy allocation requests from the program. In the context of a garbage-collected language such as Go, the use of memory arenas can offer several advantages:

* **Performance Improvement**: By allocating memory in large blocks, memory arenas reduce the overhead associated with frequent calls to the system's memory allocator. This can lead to performance improvements, especially in applications that perform many small allocations.

* **Simplified Memory Management**: In some scenarios, memory management can be simplified by allocating from an arena that is freed all at once. This is particularly useful for short-lived allocations where all the memory allocated from the arena can be released in a single operation.

* **Garbage Collection Efficiency**: Using memory arenas can reduce the workload on the garbage collector by decreasing the number of objects that need to be tracked and collected, leading to less pause time and more predictable performance.

However, while memory arenas offer these advantages, they are not a silver bullet and come with trade-offs, such as potentially increased memory usage due to unused space within the allocated blocks. Careful consideration and profiling are necessary to determine whether using a memory arena is beneficial for a particular application.

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

## Concurrency

By default, the arena implementation is not concurrent-safe, meaning it is not safe to access it concurrently from different goroutines. If the specific use case requires concurrent access, the library provides the `NewConcurrentArena` function, to which a base arena is passed and it returns a new instance that can be accessed concurrently.

```go
package main

import (
	"github.com/ortuman/nuke"
)

func main() {
	arena := nuke.NewConcurrentArena(
            nuke.NewSlabArena(256*1024, 20*1024*1024),
        )
	defer arena.Reset(true)
	
	// From here on, the arena can be safely accessed concurrently.
	// ...
}
```

## Benchmarks

Below is a comparative table with the different benchmark results.

```
BenchmarkRuntimeNewObject/100-8           	         1394955	     846.6 ns/op	     800 B/op	     100 allocs/op
BenchmarkRuntimeNewObject/1000-8          	          143031	      8357 ns/op	    8000 B/op	    1000 allocs/op
BenchmarkRuntimeNewObject/10000-8         	           14371	     83562 ns/op	   80000 B/op	   10000 allocs/op
BenchmarkRuntimeNewObject/100000-8        	            1428	    835474 ns/op	  800005 B/op	  100000 allocs/op
BenchmarkSlabArenaNewObject/100-8         	          124495	     15469 ns/op	       0 B/op	       0 allocs/op
BenchmarkSlabArenaNewObject/1000-8        	           76744	     19602 ns/op	       0 B/op	       0 allocs/op
BenchmarkSlabArenaNewObject/10000-8       	           24104	     50845 ns/op	       0 B/op	       0 allocs/op
BenchmarkSlabArenaNewObject/100000-8      	            3282	    366044 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentSlabArenaNewObject/100-8         	   90392	     16679 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentSlabArenaNewObject/1000-8        	   43753	     29823 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentSlabArenaNewObject/10000-8       	    8037	    149923 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentSlabArenaNewObject/100000-8      	     879	   1364377 ns/op	       0 B/op	       0 allocs/op
BenchmarkRuntimeMakeSlice/100-8                     	   58166	     19684 ns/op	  204800 B/op	     100 allocs/op
BenchmarkRuntimeMakeSlice/1000-8                    	    5916	    196412 ns/op	 2048010 B/op	    1000 allocs/op
BenchmarkRuntimeMakeSlice/10000-8                   	     600	   1965622 ns/op	20480106 B/op	   10001 allocs/op
BenchmarkRuntimeMakeSlice/100000-8                  	      60	  19664140 ns/op	204801155 B/op	  100012 allocs/op
BenchmarkSlabArenaMakeSlice/100-8                   	  166300	     14520 ns/op	       0 B/op	       0 allocs/op
BenchmarkSlabArenaMakeSlice/1000-8                  	   43785	     36938 ns/op	       0 B/op	       0 allocs/op
BenchmarkSlabArenaMakeSlice/10000-8                 	    2707	    427398 ns/op	       0 B/op	       0 allocs/op
BenchmarkSlabArenaMakeSlice/100000-8                	      87	  14048963 ns/op	70582284 B/op	   34464 allocs/op
BenchmarkConcurrentSlabArenaMakeSlice/100-8         	   91959	     17944 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentSlabArenaMakeSlice/1000-8        	   27384	     42790 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentSlabArenaMakeSlice/10000-8       	    2406	    480474 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentSlabArenaMakeSlice/100000-8      	      84	  14702775 ns/op	70582280 B/op	   34464 allocs/op
```

## Contributing

Contributions from the community are welcome! If you'd like to contribute, please fork the repository, make your changes, and submit a pull request.

## Contact
If you have any questions, feedback or suggestions, please feel free to contact me at ortuman@gmail.com. I'm always open to feedback and would love to hear from you!

## License

This project is licensed under the terms of the Apache-2.0 license.