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

* **Enhanced Cache Locality**: Memory arenas can also improve cache locality by allocating closely related objects within the same block of memory. This arrangement increases the likelihood that when one object is accessed, other related objects are already in the cache, thus reducing cache misses and enhancing overall application performance.

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
	// Initialize a new monotonic arena with a buffer size of 256KB 
	// and a max memory size of 20MB.
	arena := nuke.NewMonotonicArena(256*1024, 80)
	
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

	// When done, reset the arena (releasing monotonic buffer memory).
	arena.Reset(true)
	
	// From here on, any arena reference is invalid.
	// ...
}
```

Additionally, we can inject a memory arena as part of a context, with the purpose of being used throughout the lifecycle of certain operations, such as an HTTP request.

```go
func httpHandler(w http.ResponseWriter, r *http.Request) {
    // Inject memory arena into request context.
    arena := nuke.NewMonotonicArena(64*1024, 10)
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
            nuke.NewMonotonicArena(256*1024, 20),
        )
	defer arena.Reset(true)
	
	// From here on, the arena can be safely accessed concurrently.
	// ...
}
```

## Benchmarks

Below is a comparative table with the different benchmark results.

```
BenchmarkRuntimeNewObject/100-8                	                  745374	      1571 ns/op	    4800 B/op	     100 allocs/op
BenchmarkRuntimeNewObject/1000-8               	                   76626	     15633 ns/op	   48000 B/op	    1000 allocs/op
BenchmarkRuntimeNewObject/10000-8              	                    7628	    156884 ns/op	  480001 B/op	   10000 allocs/op
BenchmarkRuntimeNewObject/100000-8             	                     759	   1574775 ns/op	 4800014 B/op	  100000 allocs/op
BenchmarkRuntimeNewObject/1000000-8            	                      75	  15658095 ns/op	48000140 B/op	 1000001 allocs/op
BenchmarkMonotonicArenaNewObject/100-8         	                 1594798	     753.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkMonotonicArenaNewObject/1000-8        	                  160849	      7443 ns/op	       0 B/op	       0 allocs/op
BenchmarkMonotonicArenaNewObject/10000-8       	                   16070	     74735 ns/op	       0 B/op	       0 allocs/op
BenchmarkMonotonicArenaNewObject/100000-8      	                    1618	    745795 ns/op	       0 B/op	       0 allocs/op
BenchmarkMonotonicArenaNewObject/1000000-8     	                     146	   8097215 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentMonotonicArenaNewObject/100-8         	  848425	      1372 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentMonotonicArenaNewObject/1000-8        	   88532	     13571 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentMonotonicArenaNewObject/10000-8       	    8764	    138387 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentMonotonicArenaNewObject/100000-8      	     876	   1365637 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentMonotonicArenaNewObject/1000000-8     	      87	  13638768 ns/op	       0 B/op	       0 allocs/op
BenchmarkRuntimeMakeSlice/100-8                          	   19886	     60312 ns/op	 1024005 B/op	     100 allocs/op
BenchmarkRuntimeMakeSlice/1000-8                         	    1975	    603525 ns/op	10240051 B/op	    1000 allocs/op
BenchmarkRuntimeMakeSlice/10000-8                        	     196	   6045213 ns/op	102400519 B/op	   10005 allocs/op
BenchmarkRuntimeMakeSlice/100000-8                       	      19	  60450787 ns/op	1024005588 B/op	  100058 allocs/op
BenchmarkRuntimeMakeSlice/1000000-8                      	       2	 601334917 ns/op	10240049392 B/op	 1000514 allocs/op
BenchmarkMonotonicArenaMakeSlice/100-8                   	  147604	     11495 ns/op	       0 B/op	       0 allocs/op
BenchmarkMonotonicArenaMakeSlice/1000-8                  	    7401	    158989 ns/op	       0 B/op	       0 allocs/op
BenchmarkMonotonicArenaMakeSlice/10000-8                 	     729	   1618622 ns/op	       0 B/op	       0 allocs/op
BenchmarkMonotonicArenaMakeSlice/100000-8                	      44	  25553032 ns/op	822722688 B/op	   80345 allocs/op
BenchmarkMonotonicArenaMakeSlice/1000000-8               	       4	 274910375 ns/op	10038723976 B/op	  980358 allocs/op
BenchmarkConcurrentMonotonicArenaMakeSlice/100-8         	   61780	     19308 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentMonotonicArenaMakeSlice/1000-8        	    5998	    194522 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentMonotonicArenaMakeSlice/10000-8       	     604	   1935818 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentMonotonicArenaMakeSlice/100000-8      	      44	  25918637 ns/op	822722675 B/op	   80345 allocs/op
BenchmarkConcurrentMonotonicArenaMakeSlice/1000000-8     	       4	 276969698 ns/op	10038723640 B/op	  980355 allocs/op
```

## Contributing

Contributions from the community are welcome! If you'd like to contribute, please fork the repository, make your changes, and submit a pull request.

## Contact
If you have any questions, feedback or suggestions, please feel free to contact me at ortuman@gmail.com. I'm always open to feedback and would love to hear from you!

## License

This project is licensed under the terms of the Apache-2.0 license.
