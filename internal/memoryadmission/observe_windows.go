//go:build windows

package memoryadmission

import (
	"syscall"
	"unsafe"
)

type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

func observe() Observation {
	proc := syscall.NewLazyDLL("kernel32.dll").NewProc("GlobalMemoryStatusEx")
	status := memoryStatusEx{Length: uint32(unsafe.Sizeof(memoryStatusEx{}))}
	result, _, _ := proc.Call(uintptr(unsafe.Pointer(&status)))
	if result == 0 || status.TotalPhys == 0 || status.AvailPhys > status.TotalPhys {
		return unavailable()
	}
	const mib = uint64(1024 * 1024)
	return boundedObservation(SourceWindowsGlobal, int64(status.AvailPhys/mib), int64(status.TotalPhys/mib))
}
