//go:build !unix && !windows

package allocator

import "github.com/tetratelabs/wazero/experimental"

func alloc(cap, max uint64) experimental.LinearMemory {
	return &sliceBuffer{make([]byte, cap), max}
}

type sliceBuffer struct {
	buf []byte
	max uint64
}

func (b *sliceBuffer) Free() {}

func (b *sliceBuffer) Reallocate(size uint64) []byte {
	if cap := uint64(cap(b.buf)); size > cap {
		b.buf = append(b.buf[:cap], make([]byte, size-cap)...)
	} else {
		b.buf = b.buf[:size]
	}
	return b.buf
}
