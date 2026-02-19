package app

import (
	"testing"
)

func TestLock_Exclusive(t *testing.T) {
	tmp := t.TempDir()

	a, err := AcquireLock(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Release()

	_, err = AcquireLock(tmp)
	if err == nil {
		t.Fatalf("expected second lock attempt to fail")
	}
}
