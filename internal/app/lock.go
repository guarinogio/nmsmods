package app

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
)

type Lock struct {
	f *os.File
}

func AcquireLock(root string) (*Lock, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	lockPath := filepath.Join(root, "lock")

	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}

	// Linux/Unix: advisory flock
	if runtime.GOOS != "windows" {
		if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("another nmsmods process is running (lock busy)")
		}
	}

	return &Lock{f: f}, nil
}

func (l *Lock) Release() {
	if l == nil || l.f == nil {
		return
	}
	_ = syscall.Flock(int(l.f.Fd()), syscall.LOCK_UN)
	_ = l.f.Close()
}
