// SPDX-License-Identifier: Apache-2.0

package nuke

import (
	"fmt"
	"runtime"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/require"
)

func TestSlabArenaAllocateObject(t *testing.T) {
	arena := NewSlabArena(8182, 1) // 8KB

	var refs []*int
	for i := 0; i < 1_000; i++ {
		refs = append(refs, New[int](arena))
	}

	for i := 0; i < 1_000; i++ {
		require.True(t, isSlabArenaPtr(arena, unsafe.Pointer(refs[i])))
	}
}

func TestSlabArenaSendObjectToHeap(t *testing.T) {
	var x int
	arena := NewSlabArena(2*int(unsafe.Sizeof(x)), 1) // 2 ints room

	// Send the first two ints to the arena
	require.True(t, isSlabArenaPtr(arena, unsafe.Pointer(New[int](arena))))
	require.True(t, isSlabArenaPtr(arena, unsafe.Pointer(New[int](arena))))

	// Send last one to the heap
	require.False(t, isSlabArenaPtr(arena, unsafe.Pointer(New[int](arena))))
}

func TestSlabArenaReset(t *testing.T) {
	arena := NewSlabArena(1024, 1).(*slabArena) // one slab of 1KB

	// Allocate slab buffer
	_ = New[int](arena)

	// Configure finalizer
	gced := make(chan bool)
	runtime.SetFinalizer((*byte)(arena.slabs[0].ptr), func(*byte) {
		close(gced)
	})

	// Reset the arena (without releasing memory)
	arena.Reset(false)
	runtime.GC()

	select {
	case <-gced:
		require.Fail(t, "finalizer should not have been called")

	case <-time.NewTimer(time.Second).C:
		break
	}

	// Do another allocation
	_ = New[int](arena)

	// Reset the arena (releasing memory)
	arena.Reset(true)
	runtime.GC()

	select {
	case <-gced:
		break

	case <-time.NewTimer(time.Second).C:
		require.Fail(t, "finalizer should have been called")
	}

	// Add this extra allocation here to prevent the compiler from marking arena reference as unused
	// before invoking runtime.GC().
	_ = New[int](arena)
}

func TestSlabArenaAllocateSlice(t *testing.T) {}

func isSlabArenaPtr(a Arena, ptr unsafe.Pointer) bool {
	sa := a.(*slabArena)
	for _, s := range sa.slabs {
		if s.ptr == nil {
			break
		}
		beginPtr := uintptr(s.ptr)
		endPtr := uintptr(s.ptr) + uintptr(s.size)

		if uintptr(ptr) >= beginPtr && uintptr(ptr) < endPtr {
			return true
		}
	}
	return false
}

func BenchmarkRuntimeNewObject(b *testing.B) {
	a := newRuntimeAllocator[int]()
	for _, objectCount := range []int{100, 1_000, 10_000, 100_000} {
		b.Run(fmt.Sprintf("%d", objectCount), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				for j := 0; j < objectCount; j++ {
					_ = a.new()
				}
			}
		})
	}
}

func BenchmarkSlabArenaNewObject(b *testing.B) {
	slabArena := NewSlabArena(1024*1024, 128) // 1Mb slab size (128 MB)

	a := newArenaAllocator[int](slabArena)
	for _, objectCount := range []int{100, 1_000, 10_000, 100_000} {
		b.Run(fmt.Sprintf("%d", objectCount), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				for j := 0; j < objectCount; j++ {
					_ = a.new()
				}
				a.(*arenaAllocator[int]).a.Reset(false)
			}
		})
	}
}

func BenchmarkConcurrentSlabArenaNewObject(b *testing.B) {
	slabArena := NewSlabArena(1024*1024, 128) // 1Mb slab size (128 MB)

	a := newArenaAllocator[int](NewConcurrentArena(slabArena))
	for _, objectCount := range []int{100, 1_000, 10_000, 100_000} {
		b.Run(fmt.Sprintf("%d", objectCount), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				for j := 0; j < objectCount; j++ {
					_ = a.new()
				}
				a.(*arenaAllocator[int]).a.Reset(false)
			}
		})
	}
}

func BenchmarkRuntimeMakeSlice(b *testing.B) {
	a := newRuntimeAllocator[int]()
	for _, objectCount := range []int{100, 1_000, 10_000, 100_000} {
		b.Run(fmt.Sprintf("%d", objectCount), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				for j := 0; j < objectCount; j++ {
					_ = a.makeSlice(0, 256)
				}
			}
		})
	}
}

func BenchmarkSlabArenaMakeSlice(b *testing.B) {
	slabArena := NewSlabArena(1024*1024, 128) // 1Mb slab size (128 MB)

	a := newArenaAllocator[int](slabArena)
	for _, objectCount := range []int{100, 1_000, 10_000, 100_000} {
		b.Run(fmt.Sprintf("%d", objectCount), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				for j := 0; j < objectCount; j++ {
					_ = a.makeSlice(0, 256)
				}
				a.(*arenaAllocator[int]).a.Reset(false)
			}
		})
	}
}

func BenchmarkConcurrentSlabArenaMakeSlice(b *testing.B) {
	slabArena := NewSlabArena(1024*1024, 128) // 1Mb slab size (128 MB)

	a := newArenaAllocator[int](NewConcurrentArena(slabArena))
	for _, objectCount := range []int{100, 1_000, 10_000, 100_000} {
		b.Run(fmt.Sprintf("%d", objectCount), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				for j := 0; j < objectCount; j++ {
					_ = a.makeSlice(0, 256)
				}
				a.(*arenaAllocator[int]).a.Reset(false)
			}
		})
	}
}

type allocator[T any] interface {
	new() *T
	makeSlice(len, cap int) []T
}

type runtimeAllocator[T any] struct{}

func newRuntimeAllocator[T any]() allocator[T] {
	return &runtimeAllocator[T]{}
}

func (r *runtimeAllocator[T]) new() *T                    { return new(T) }
func (r *runtimeAllocator[T]) makeSlice(len, cap int) []T { return make([]T, len, cap) }

type arenaAllocator[T any] struct {
	a Arena
}

func newArenaAllocator[T any](a Arena) allocator[T] {
	return &arenaAllocator[T]{a: a}
}

func (r *arenaAllocator[T]) new() *T                    { return New[T](r.a) }
func (r *arenaAllocator[T]) makeSlice(len, cap int) []T { return MakeSlice[T](r.a, len, cap) }
