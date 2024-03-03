// SPDX-License-Identifier: Apache-2.0

package nuke

import (
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
