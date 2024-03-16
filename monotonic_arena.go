// SPDX-License-Identifier: Apache-2.0

package nuke

import (
	"unsafe"
)

type monotonicArena struct {
	buffers []*monotonicBuffer
}

type monotonicBuffer struct {
	ptr    unsafe.Pointer
	offset uintptr
	size   uintptr
}

func newMonotonicBuffer(size int) *monotonicBuffer {
	return &monotonicBuffer{size: uintptr(size)}
}

func (s *monotonicBuffer) alloc(size, alignment uintptr) (unsafe.Pointer, bool) {
	if s.ptr == nil {
		buf := make([]byte, s.size) // allocate monotonic buffer lazily
		s.ptr = unsafe.Pointer(unsafe.SliceData(buf))
	}
	alignOffset := uintptr(0)
	if delta := (uintptr(s.ptr) + s.offset) % alignment; delta > 0 {
		alignOffset = alignment - delta
	}
	allocSize := size + alignOffset

	if s.availableBytes() < allocSize {
		return nil, false
	}
	ptr := unsafe.Pointer(uintptr(s.ptr) + s.offset + alignOffset)
	s.offset += allocSize

	clear(unsafe.Slice((*byte)(ptr), size))

	return ptr, true
}

func (s *monotonicBuffer) reset(release bool) {
	if s.offset == 0 {
		return
	}
	s.offset = 0

	if release {
		s.ptr = nil
	}
}

func (s *monotonicBuffer) availableBytes() uintptr {
	return s.size - s.offset
}

// NewMonotonicArena creates a new monotonic arena with a specified number of buffers and a buffer size.
func NewMonotonicArena(bufferSize, bufferCount int) Arena {
	a := &monotonicArena{buffers: make([]*monotonicBuffer, 0, bufferCount)}
	for i := 0; i < bufferCount; i++ {
		a.buffers = append(a.buffers, newMonotonicBuffer(bufferSize))
	}
	return a
}

// Alloc satisfies the Arena interface.
func (a *monotonicArena) Alloc(size, alignment uintptr) unsafe.Pointer {
	for i := range a.buffers {
		ptr, ok := a.buffers[i].alloc(size, alignment)
		if ok {
			return ptr
		}
	}
	return nil
}

// Reset satisfies the Arena interface.
func (a *monotonicArena) Reset(release bool) {
	for i := range a.buffers {
		a.buffers[i].reset(release)
	}
}
