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
BenchmarkRuntimeNewObject/100-8                	                 1310493	     909.2 ns/op	     800 B/op	     100 allocs/op
BenchmarkRuntimeNewObject/1000-8               	                  132147	      9030 ns/op	    8000 B/op	    1000 allocs/op
BenchmarkRuntimeNewObject/10000-8              	                   13296	     90249 ns/op	   80000 B/op	   10000 allocs/op
BenchmarkRuntimeNewObject/100000-8             	                    1316	    904476 ns/op	  800004 B/op	  100000 allocs/op
BenchmarkRuntimeNewObject/1000000-8            	                     130	   9033261 ns/op	 8000044 B/op	 1000000 allocs/op
BenchmarkMonotonicArenaNewObject/100-8         	                 2266246	     530.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkMonotonicArenaNewObject/1000-8        	                  228908	      5200 ns/op	       0 B/op	       0 allocs/op
BenchmarkMonotonicArenaNewObject/10000-8       	                   23200	     51807 ns/op	       0 B/op	       0 allocs/op
BenchmarkMonotonicArenaNewObject/100000-8      	                    2312	    519789 ns/op	       0 B/op	       0 allocs/op
BenchmarkMonotonicArenaNewObject/1000000-8     	                     229	   5203328 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentMonotonicArenaNewObject/100-8         	  884904	      1357 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentMonotonicArenaNewObject/1000-8        	   88495	     13526 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentMonotonicArenaNewObject/10000-8       	    8844	    135562 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentMonotonicArenaNewObject/100000-8      	     885	   1359547 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentMonotonicArenaNewObject/1000000-8     	      87	  13591854 ns/op	       0 B/op	       0 allocs/op
BenchmarkRuntimeMakeSlice/100-8                          	   57231	     20169 ns/op	  204800 B/op	     100 allocs/op
BenchmarkRuntimeMakeSlice/1000-8                         	    5780	    206335 ns/op	 2048007 B/op	    1000 allocs/op
BenchmarkRuntimeMakeSlice/10000-8                        	     585	   2017157 ns/op	20480080 B/op	   10000 allocs/op
BenchmarkRuntimeMakeSlice/100000-8                       	      57	  20167039 ns/op	204800759 B/op	  100007 allocs/op
BenchmarkRuntimeMakeSlice/1000000-8                      	       5	 200384042 ns/op	2048007507 B/op	 1000078 allocs/op
BenchmarkMonotonicArenaMakeSlice/100-8                   	  627627	      2219 ns/op	       0 B/op	       0 allocs/op
BenchmarkMonotonicArenaMakeSlice/1000-8                  	   52328	     22791 ns/op	       0 B/op	       0 allocs/op
BenchmarkMonotonicArenaMakeSlice/10000-8                 	    3667	    312075 ns/op	       0 B/op	       0 allocs/op
BenchmarkMonotonicArenaMakeSlice/100000-8                	     164	   6971221 ns/op	70582281 B/op	   34464 allocs/op
BenchmarkMonotonicArenaMakeSlice/1000000-8               	       8	 128829224 ns/op	1913782512 B/op	  934466 allocs/op
BenchmarkConcurrentMonotonicArenaMakeSlice/100-8         	  157754	      8651 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentMonotonicArenaMakeSlice/1000-8        	   13676	     87545 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentMonotonicArenaMakeSlice/10000-8       	    1358	    879413 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentMonotonicArenaMakeSlice/100000-8      	     109	  10689078 ns/op	70582276 B/op	   34464 allocs/op
BenchmarkConcurrentMonotonicArenaMakeSlice/1000000-8     	       8	 133404771 ns/op	1913782680 B/op	  934468 allocs/op
```

## Contributing

Contributions from the community are welcome! If you'd like to contribute, please fork the repository, make your changes, and submit a pull request.

## Contact
If you have any questions, feedback or suggestions, please feel free to contact me at ortuman@gmail.com. I'm always open to feedback and would love to hear from you!

## License

This project is licensed under the terms of the Apache-2.0 license.