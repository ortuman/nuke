// SPDX-License-Identifier: Apache-2.0

package nuke

const growThreshold = 256

// SliceAppend appends elements to a slice of type T using a provided Arena
// for memory allocation if needed.
func SliceAppend[T any](a Arena, s []T, data ...T) []T {
	if a == nil {
		return append(s, data...)
	}
	s = growSlice(a, s, len(data))
	s = append(s, data...)
	return s
}

func growSlice[T any](a Arena, s []T, dataLen int) []T {
	newLen := len(s) + dataLen
	newCap := cap(s)

	if newCap > 0 {
		for newLen > newCap {
			if newCap < growThreshold {
				newCap *= 2
			} else {
				newCap += newCap / 4
			}
		}
	} else {
		newCap = dataLen
	}
	if newCap == cap(s) {
		return s
	}
	s2 := MakeSlice[T](a, len(s), newCap)
	copy(s2, s)
	return s2
}
