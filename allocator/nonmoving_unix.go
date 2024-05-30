//go:build unix

package allocator

import (
	"fmt"
	"math"
	"syscall"

	"github.com/tetratelabs/wazero/experimental"
)

func alloc(_, max uint64) experimental.LinearMemory {
	// Round up to the page size.
	rnd := uint64(syscall.Getpagesize() - 1)
	max = (max + rnd) &^ rnd

	if max > math.MaxInt {
		// This ensures int(max) overflows to a negative value,
		// and syscall.Mmap returns EINVAL.
		max = math.MaxUint64
	}

	// Reserve max bytes of address space, to ensure we won't need to move it.
	// A protected, private, anonymous mapping should not commit memory.
	b, err := syscall.Mmap(-1, 0, int(max), syscall.PROT_NONE, syscall.MAP_PRIVATE|syscall.MAP_ANON)
	if err != nil {
		panic(fmt.Errorf("allocator_unix: failed to reserve memory: %w", err))
	}
	return &mmappedMemory{buf: b[:0]}
}

// The slice covers the entire mmapped memory:
//   - len(buf) is the already committed memory,
//   - cap(buf) is the reserved address space.
type mmappedMemory struct {
	buf []byte
}

func (m *mmappedMemory) Reallocate(size uint64) []byte {
	com := uint64(len(m.buf))
	res := uint64(cap(m.buf))
	if com < size && size < res {
		// Round up to the page size.
		rnd := uint64(syscall.Getpagesize() - 1)
		new := (size + rnd) &^ rnd

		// Commit additional memory up to new bytes.
		err := syscall.Mprotect(m.buf[com:new], syscall.PROT_READ|syscall.PROT_WRITE)
		if err != nil {
			panic(fmt.Errorf("allocator_unix: failed to commit memory: %w", err))
		}

		// Update committed memory.
		m.buf = m.buf[:new]
	}
	// Limit returned capacity because bytes beyond
	// len(m.buf) have not yet been committed.
	return m.buf[:size:len(m.buf)]
}

func (m *mmappedMemory) Free() {
	err := syscall.Munmap(m.buf[:cap(m.buf)])
	if err != nil {
		panic(fmt.Errorf("allocator_unix: failed to release memory: %w", err))
	}
	m.buf = nil
}
