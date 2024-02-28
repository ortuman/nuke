// SPDX-License-Identifier: Apache-2.0

package nuke

import (
	"sync"
	"unsafe"
)

type slabArena struct {
	slabs []*slab
}

type slab struct {
	mtx    sync.Mutex
	ptr    unsafe.Pointer
	offset int
	size   int
}

func newSlab(size int) *slab {
	return &slab{
		size: size,
	}
}

func (s *slab) alloc(size int) (unsafe.Pointer, bool) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.ptr == nil {
		buf := make([]byte, s.size) // allocate slab buffer lazily
		s.ptr = unsafe.Pointer(unsafe.SliceData(buf))
	}
	if s.availableBytes() < size {
		return nil, false
	}
	ptr := unsafe.Pointer(uintptr(s.ptr) + uintptr(s.offset))
	s.offset += size

	return ptr, true
}

func (s *slab) reset(release bool) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.offset == 0 {
		return
	}
	s.offset = 0

	if release {
		s.ptr = nil
	} else {
		s.zeroOutBuffer()
	}
}

func (s *slab) zeroOutBuffer() {
	b := unsafe.Slice((*byte)(s.ptr), s.size)

	// This piece of code will be translated into a runtime.memclrNoHeapPointers
	// invocation by the compiler, which is an assembler optimized implementation.
	// Architecture specific code can be found at src/runtime/memclr_$GOARCH.s
	// in Go source (since https://codereview.appspot.com/137880043).
	for i := range b {
		b[i] = 0
	}
}

func (s *slab) availableBytes() int {
	return s.size - s.offset
}

// NewSlabArena creates a new slab arena with a specified number of slabs and slab size.
func NewSlabArena(slabSize, slabCount int) Arena {
	a := &slabArena{}
	for i := 0; i < slabCount; i++ {
		a.slabs = append(a.slabs, newSlab(slabSize))
	}
	return a
}

// Alloc satisfies the Arena interface.
func (a *slabArena) Alloc(size int) unsafe.Pointer {
	for i := 0; i < len(a.slabs); i++ {
		ptr, ok := a.slabs[i].alloc(size)
		if ok {
			return ptr
		}
	}
	return nil
}

// Reset satisfies the Arena interface.
func (a *slabArena) Reset(release bool) {
	for _, s := range a.slabs {
		s.reset(release)
	}
}
