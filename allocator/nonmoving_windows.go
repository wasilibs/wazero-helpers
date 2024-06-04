//go:build windows

package allocator

import (
	"fmt"
	"math"
	"sync"
	"syscall"
	"unsafe"

	"github.com/tetratelabs/wazero/experimental"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procVirtualAlloc = kernel32.NewProc("VirtualAlloc")
	procVirtualFree  = kernel32.NewProc("VirtualFree")
)

const (
	windowsMemCommit     uintptr = 0x00001000
	windowsMemReserve    uintptr = 0x00002000
	windowsMemRelease    uintptr = 0x00008000
	windowsPageReadwrite uintptr = 0x00000004

	// https://cs.opensource.google/go/x/sys/+/refs/tags/v0.20.0:windows/syscall_windows.go;l=131
	pageSize = 4096
)

func alloc(_, max uint64) experimental.LinearMemory {
	// Round up to the page size because recommitting must be page-aligned.
	// In practice, the WebAssembly page size should be a multiple of the system
	// page size on most if not all platforms and rounding will never happen.
	rnd := uint64(pageSize) - 1
	reserved := (max + rnd) &^ rnd

	if reserved > math.MaxInt {
		// This ensures uintptr(max) overflows to a large value,
		// and windows.VirtualAlloc returns an error.
		reserved = math.MaxUint64
	}

	// Reserve max bytes of address space, to ensure we won't need to move it.
	// This does not commit memory.
	r, _, err := procVirtualAlloc.Call(0, uintptr(reserved), windowsMemReserve, windowsPageReadwrite)
	if r == 0 {
		panic(fmt.Errorf("allocator_windows: failed to reserve memory: %w", err))
	}

	buf := unsafe.Slice((*byte)(unsafe.Pointer(r)), int(reserved))
	return &virtualMemory{buf: buf[:0], addr: r, max: max}
}

// The slice covers the entire mmapped memory:
//   - len(buf) is the already committed memory,
//   - cap(buf) is the reserved address space, which is max rounded up to a page.
type virtualMemory struct {
	buf  []byte
	addr uintptr
	max  uint64

	// Any reasonable Wasm implementation will take a lock before calling Grow, but this
	// is invisible to Go's race detector so it can still detect raciness when we updated
	// buf. We go ahead and take a lock when mutating since the performance effect should
	// be negligible in practice and it will help the race detector confirm the safety.
	mu sync.Mutex
}

func (m *virtualMemory) Reallocate(size uint64) []byte {
	if size > m.max {
		panic(errInvalidReallocation)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	com := uint64(len(m.buf))
	if com < size {
		// Round up to the page size.
		rnd := uint64(pageSize) - 1
		newCap := (size + rnd) &^ rnd

		// Commit additional memory up to new bytes.
		r, _, err := procVirtualAlloc.Call(m.addr, uintptr(newCap), windowsMemCommit, windowsPageReadwrite)
		if r == 0 {
			panic(fmt.Errorf("allocator_windows: failed to commit memory: %w", err))
		}

		// Update committed memory.
		m.buf = m.buf[:newCap]
	}
	// Limit returned capacity because bytes beyond
	// len(m.buf) have not yet been committed.
	return m.buf[:size:len(m.buf)]
}

func (m *virtualMemory) Free() {
	m.mu.Lock()
	defer m.mu.Unlock()

	r, _, err := procVirtualFree.Call(m.addr, 0, windowsMemRelease)
	if r == 0 {
		panic(fmt.Errorf("allocator_windows: failed to release memory: %w", err))
	}
	m.addr = 0
}
