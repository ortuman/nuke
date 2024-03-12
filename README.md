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
BenchmarkRuntimeNewObject/100-8                	                 1000000	      1009 ns/op	     800 B/op	     100 allocs/op
BenchmarkRuntimeNewObject/1000-8               	                  117721	      9963 ns/op	    8000 B/op	    1000 allocs/op
BenchmarkRuntimeNewObject/10000-8              	                   10000	    100035 ns/op	   80000 B/op	   10000 allocs/op
BenchmarkRuntimeNewObject/100000-8             	                    1202	   1003802 ns/op	  800008 B/op	  100000 allocs/op
BenchmarkRuntimeNewObject/1000000-8            	                     120	  10089223 ns/op	 8000086 B/op	 1000000 allocs/op
BenchmarkMonotonicArenaNewObject/100-8         	                 1000000	      2810 ns/op	     800 B/op	      86 allocs/op
BenchmarkMonotonicArenaNewObject/1000-8        	                   88327	     27757 ns/op	    8000 B/op	     851 allocs/op
BenchmarkMonotonicArenaNewObject/10000-8       	                   10000	    278321 ns/op	   80000 B/op	    8689 allocs/op
BenchmarkMonotonicArenaNewObject/100000-8      	                     922	   2776261 ns/op	  800000 B/op	   85783 allocs/op
BenchmarkMonotonicArenaNewObject/1000000-8     	                     100	  27928883 ns/op	 8000001 B/op	  868928 allocs/op
BenchmarkConcurrentMonotonicArenaNewObject/100-8         	    5194	    223815 ns/op	    2018 B/op	       0 allocs/op
BenchmarkConcurrentMonotonicArenaNewObject/1000-8        	    4696	    230037 ns/op	    2232 B/op	       0 allocs/op
BenchmarkConcurrentMonotonicArenaNewObject/10000-8       	    3408	    345998 ns/op	    3076 B/op	       0 allocs/op
BenchmarkConcurrentMonotonicArenaNewObject/100000-8      	     729	   1618922 ns/op	   14383 B/op	       0 allocs/op
BenchmarkConcurrentMonotonicArenaNewObject/1000000-8     	      81	  14709630 ns/op	  129453 B/op	       0 allocs/op
BenchmarkRuntimeMakeSlice/100-8                          	   47329	     25681 ns/op	  204800 B/op	     100 allocs/op
BenchmarkRuntimeMakeSlice/1000-8                         	    4420	    261322 ns/op	 2048010 B/op	    1000 allocs/op
BenchmarkRuntimeMakeSlice/10000-8                        	     469	   2525088 ns/op	20480097 B/op	   10001 allocs/op
BenchmarkRuntimeMakeSlice/100000-8                       	      46	  24549537 ns/op	204800932 B/op	  100009 allocs/op
BenchmarkRuntimeMakeSlice/1000000-8                      	       4	 259382260 ns/op	2048009360 B/op	 1000097 allocs/op
BenchmarkMonotonicArenaMakeSlice/100-8                   	   67718	     18072 ns/op	  204800 B/op	      99 allocs/op
BenchmarkMonotonicArenaMakeSlice/1000-8                  	    8508	    181105 ns/op	 2048000 B/op	     993 allocs/op
BenchmarkMonotonicArenaMakeSlice/10000-8                 	     720	   1709450 ns/op	20480004 B/op	    9928 allocs/op
BenchmarkMonotonicArenaMakeSlice/100000-8                	      92	  17137978 ns/op	204800060 B/op	   99444 allocs/op
BenchmarkMonotonicArenaMakeSlice/1000000-8               	       6	 175183750 ns/op	2048000576 B/op	  991474 allocs/op
BenchmarkConcurrentMonotonicArenaMakeSlice/100-8         	   69325	     17460 ns/op	  204800 B/op	      99 allocs/op
BenchmarkConcurrentMonotonicArenaMakeSlice/1000-8        	    8067	    175321 ns/op	 2048000 B/op	     993 allocs/op
BenchmarkConcurrentMonotonicArenaMakeSlice/10000-8       	     696	   1767369 ns/op	20480012 B/op	    9926 allocs/op
BenchmarkConcurrentMonotonicArenaMakeSlice/100000-8      	      98	  17512536 ns/op	204800064 B/op	   99478 allocs/op
BenchmarkConcurrentMonotonicArenaMakeSlice/1000000-8     	       6	 174489826 ns/op	2048000624 B/op	  991474 allocs/op
```

## Contributing

Contributions from the community are welcome! If you'd like to contribute, please fork the repository, make your changes, and submit a pull request.

## Contact
If you have any questions, feedback or suggestions, please feel free to contact me at ortuman@gmail.com. I'm always open to feedback and would love to hear from you!

## License

This project is licensed under the terms of the Apache-2.0 license.