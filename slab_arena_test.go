// SPDX-License-Identifier: Apache-2.0

package nuke

import (
	"testing"
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

func TestSlabArenaAllocateSlice(t *testing.T) {}

func isSlabArenaPtr(a Arena, ptr unsafe.Pointer) bool {
	sa := a.(*slabArena)
	for _, s := range sa.slabs {
		beginPtr := uintptr(s.ptr)
		endPtr := uintptr(s.ptr) + uintptr(s.size)

		if uintptr(ptr) >= beginPtr && uintptr(ptr) < endPtr {
			return true
		}
	}
	return false
}
