// SPDX-License-Identifier: Apache-2.0

package nuke

import (
	"sync"
	"unsafe"
)

type concurrentArena struct {
	mtx sync.Mutex
	a   Arena
}

// NewConcurrentArena returns an arena that is safe to be accessed concurrently
// from multiple goroutines.
func NewConcurrentArena(a Arena) Arena {
	return &concurrentArena{a: a}
}

// Alloc satisfies the Arena interface.
func (a *concurrentArena) Alloc(size, alignment uintptr) unsafe.Pointer {
	a.mtx.Lock()
	ptr := a.a.Alloc(size, alignment)
	a.mtx.Unlock()
	return ptr
}

// Reset satisfies the Arena interface.
func (a *concurrentArena) Reset(release bool) {
	a.mtx.Lock()
	a.a.Reset(release)
	a.mtx.Unlock()
}
