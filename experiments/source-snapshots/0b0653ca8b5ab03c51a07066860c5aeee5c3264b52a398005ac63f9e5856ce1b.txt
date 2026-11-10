//go:build darwin && liftrs_cpp && cgo

package main

import (
	"os"
	"syscall"
)

// F_NOCACHE=48 is defined by the local macOS SDK sys/fcntl.h. This asks the
// kernel to bypass its data cache for this fd, not the SSD/controller caches.
func noCache(f *os.File) (string, error) {
	_, _, err := syscall.Syscall(syscall.SYS_FCNTL, f.Fd(), 48, 1)
	if err != 0 {
		return "", err
	}
	return "macOS F_NOCACHE", nil
}
